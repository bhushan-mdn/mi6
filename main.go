package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Agent struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Port   string `json:"port"`
	Status string `json:"status"`
}

var db *sql.DB

type AgentMgrCmd struct {
	ServerId int
	Command  string
}

func AgentManager(inCh chan AgentMgrCmd, outCh chan string) {
	for inMsg := range inCh {
		switch inMsg.Command {
		case "start":
			fmt.Println("starting server", inMsg.ServerId)
			outCh <- "started"
		case "stop":
			fmt.Println("stopping server", inMsg.ServerId)
			outCh <- "stopped"
		}
	}
}

func startAllAgents(w http.ResponseWriter, r *http.Request) {

}

func main() {
	port := flag.String("port", "6969", "port to run the server on")
	flag.Parse()

	var err error
	db, err = sql.Open("sqlite3", "agents.db")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/agents", func(r chi.Router) {
		r.Get("/", listAgents)
		r.Post("/", createAgent)
		r.Post("/startAll", startAllAgents)

		r.Route("/{agentID}", func(r chi.Router) {
			r.Use(AgentCtx)
			r.Get("/", getAgent)
			r.Post("/start", startAgent)
		})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", *port),
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("starting server on :%s", *port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit

	log.Println("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// This is where the graceful shutdown happens
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server stopped")
}

func AgentCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agentID := chi.URLParam(r, "agentID")
		agent, err := dbGetAgent(r.Context(), agentID)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "agent", agent)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var ErrAgentNotFound = fmt.Errorf("Agent not found")

func dbGetAgent(ctx context.Context, agentID string) (*Agent, error) {
	var agent Agent
	agentIDInt, err := strconv.Atoi(agentID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	row := db.QueryRowContext(ctx, "SELECT id, name, port, status FROM agents WHERE id = ?", agentIDInt)
	if err := row.Scan(&agent.Id, &agent.Name, &agent.Port, &agent.Status); err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(agent)

	return &agent, nil
}

func dbUpdateAgent(ctx context.Context, agentID string, status string) error {
	var agent Agent
	agentIDInt, err := strconv.Atoi(agentID)
	if err != nil {
		log.Println(err)
		return err
	}
	stmt, _ := db.Prepare("UPDATE agents SET status = ? WHERE id = ?")

	result, err := stmt.ExecContext(ctx, status, agentIDInt)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(result.RowsAffected())
	log.Println(agent)

	return nil
}

func getAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agent, ok := ctx.Value("agent").(*Agent)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	w.Write(fmt.Appendf([]byte(""), "name:%s", agent.Name))
}

func startAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agent, ok := ctx.Value("agent").(*Agent)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	rows, err := db.QueryContext(ctx, "SELECT id, path, response FROM agent_paths")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(404), 404)
		return
	}
	defer rows.Close()

	var agentPaths []AgentPath
	for rows.Next() {
		var agentPath AgentPath

		if err := rows.Scan(&agentPath.Id, &agentPath.Path, &agentPath.Response); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		agentPaths = append(agentPaths, agentPath)

	}

	mux := http.NewServeMux()

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", agent.Port),
		Handler: mux,
	}
	for _, a := range agentPaths {
		mux.HandleFunc(a.Path, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(a.Response))
		})
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}

	err = dbUpdateAgent(ctx, strconv.Itoa(agent.Id), "active")
	if err != nil {
		log.Println(err)
	}
}

type AgentPath struct {
	Id       int
	Path     string
	Response string
}

func listAgents(w http.ResponseWriter, r *http.Request) {
	var agents []Agent

	rows, err := db.QueryContext(r.Context(), "SELECT id, name, port, status FROM agents")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var agent Agent
		if err := rows.Scan(&agent.Id, &agent.Name, &agent.Port, &agent.Status); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		agents = append(agents, agent)
	}

	agentsBytes, err := json.Marshal(&agents)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Write(agentsBytes)
}

func createAgent(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))

	// stmt, err := db.Prepare("INSERT INTO agents(id, name, port) VALUES (?, ?, ?)")
	// must(err)
	// _, err = stmt.Exec(1, "svc1", "8080")
	// must(err)
}
