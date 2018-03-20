package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/goraft/raft"
	//"github.com/goraft/raftd/command"
	//"github.com/goraft/raftd/db"
	"github.com/gorilla/mux"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"sync"
	"time"
	"strings"
	"flag"
	"os"
	"net"
	"../pb"
)

// The raftd server is a combination of the Raft server and an HTTP
// server which acts as the transport.
type Server struct {
	name       string
	host       string
	port       int
	path       string
	masternode string
	router     *mux.Router
	raftServer raft.Server
	httpServer *http.Server
	//db         *db.DB
	mutex      sync.RWMutex
	workers   map[string]bool
}
type ServerOption struct {

	verbose      bool
    trace        bool
    debug        bool
    host         string
    port         int
    join         string
}
type ClusterStatusResult struct {
	IsLeader bool     `json:"IsLeader,omitempty"`
	Leader   string   `json:"Leader,omitempty"`
	Peers    []string `json:"Peers,omitempty"`
}
var (
  Soption    ServerOption
  SFlag      flag.FlagSet
)

func init() {
	SFlag.BoolVar(&Soption.verbose, "v", false, "verbose logging")
	SFlag.BoolVar(&Soption.trace, "trace", false, "Raft trace debugging")
	SFlag.BoolVar(&Soption.debug, "debug", false, "Raft debugging")
	SFlag.StringVar(&Soption.host, "h", "localhost", "hostname")
	SFlag.IntVar(&Soption.port, "p", 4001, "port")
	SFlag.StringVar(&Soption.join, "join", "", "host:port of leader to join")
	SFlag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments] <data-path> \n", os.Args[0])
		SFlag.PrintDefaults()
	}
}
// Creates a new server.
func NewServer(path string, host string, port int,masternode string) *Server {
	s := &Server{
		host:   host,
		port:   port,
		path:   path,
		//db:     db.New(),
		masternode:masternode,
		router: mux.NewRouter(),
		workers:   make(map[string]bool),
	}

	// Read existing name or generate a new one.
	if b, err := ioutil.ReadFile(filepath.Join(path, "name")); err == nil {
		s.name = string(b)
	} else {
		s.name = fmt.Sprintf("%07x", rand.Int())[0:7]
		if err = ioutil.WriteFile(filepath.Join(path, "name"), []byte(s.name), 0644); err != nil {
			panic(err)
		}
	}

	return s
}
func RunServer(argvs []string){
	
    SFlag.Parse(argvs)
    //args := flag.Args()

	if Soption.verbose {
		log.Print("Verbose logging enabled.")
	}
	if Soption.trace {
		raft.SetLogLevel(raft.Trace)
		log.Print("Raft trace debugging enabled.")
	} else if Soption.debug {
		raft.SetLogLevel(raft.Debug)
		log.Print("Raft debugging enabled.")
	}

    if SFlag.NArg() == 0 {
		SFlag.Usage()
		log.Fatal("Data path argument required")
	}
	path := SFlag.Arg(0)
	if err := os.MkdirAll(path, 0744); err != nil {
		log.Fatalf("Unable to create path: %v", err)
	} 
	var masternode string 
    if Soption.join == ""{
    	masternode= fmt.Sprintf("%s:%d",Soption.host,Soption.port)
    }else{
    	masternode=Soption.join
    }
    server := NewServer(path, Soption.host, Soption.port,masternode)
  	server.ListenAndServe(Soption.join)
	//log.Fatal(server.ListenAndServe(Soption.join))
}
// Returns the connection string.
func (s *Server) connectionString() string {
	return fmt.Sprintf("http://%s:%d", s.host, s.port)
}

// Starts the server.
func (s *Server) ListenAndServe(leader string) error {
	var err error

	log.Printf("Initializing Raft Server: %s,host:%s,port:%d", s.path,Soption.host,Soption.port)

	// Initialize and start Raft server.
	transporter := raft.NewHTTPTransporter("/raft", 200*time.Millisecond)
	s.raftServer, err = raft.NewServer(s.name, s.path, transporter, nil, nil, "")
	if err != nil {
		log.Fatal(err)
	}
	transporter.Install(s.raftServer, s)
	s.raftServer.Start()

	if leader != "" {
		// Join to leader if specified.

		log.Println("Attempting to join leader:", leader)

		if !s.raftServer.IsLogEmpty() {
			log.Fatal("Cannot join with an existing log")
		}
		if err := s.Join(leader); err != nil {
			log.Fatal(err)
		}

	} else if s.raftServer.IsLogEmpty() {
		// Initialize the server by joining itself.

		log.Println("Initializing new cluster")

		_, err := s.raftServer.Do(&raft.DefaultJoinCommand{
			Name:             s.raftServer.Name(),
			ConnectionString: s.connectionString(),
		})
		if err != nil {
			log.Fatal(err)
		}

	} else {
		log.Println("Recovered from log")
	}

	log.Println("Initializing HTTP server")

	
	//s.router.HandleFunc("/db/{key}", s.readHandler).Methods("GET")
	//s.router.HandleFunc("/db/{key}", s.writeHandler).Methods("POST")
	s.router.HandleFunc("/cluster/join", s.joinHandler).Methods("POST")
	s.router.HandleFunc("/cluster/status", s.statusHandler).Methods("GET")
	s.router.HandleFunc("/worker/status", s.workerHandler).Methods("GET")


    //lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d",Soption.host,4002))


	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d",Soption.host,4001))
	if err != nil {
	        log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Listening at:", s.connectionString())

	m := cmux.New(lis)
    
    grpcL := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpL := m.Match(cmux.Any())

	grpcServer := grpc.NewServer()

	pb.RegisterMsgServer(grpcServer, s)

    go grpcServer.Serve(grpcL)


    httpS := &http.Server{Handler: s.router}
    go httpS.Serve(httpL)
    go m.Serve()
    /*if err := m.Serve(); err != nil {
		log.Fatalf("master server failed to serve", err)
	}*/

	return nil
	
}

// This is a hack around Gorilla mux not providing the correct net/http
// HandleFunc() interface.
func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.router.HandleFunc(pattern, handler)
}

// Joins to the leader of an existing cluster.
func (s *Server) Join(leader string) error {
	command := &raft.DefaultJoinCommand{
		Name:             s.raftServer.Name(),
		ConnectionString: s.connectionString(),
	}

	var b bytes.Buffer
	json.NewEncoder(&b).Encode(command)
	resp, err := http.Post(fmt.Sprintf("http://%s/cluster/join", leader), "application/json", &b)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func (s *Server) joinHandler(w http.ResponseWriter, req *http.Request) {
	command := &raft.DefaultJoinCommand{}

	if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := s.raftServer.Do(command); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (s *Server) Peers() (members []string) {
	peers := s.raftServer.Peers()

	for _, p := range peers {
		members = append(members, strings.TrimPrefix(p.ConnectionString, "http://"))
	}

	return
}
func (s Server) SendHeartbeat(stream pb.Msg_SendHeartbeatServer) error {
	 var worknode string 
     for{
        heartbeat, err := stream.Recv()
        if err!=nil{
        	s.deleteWorker(worknode)
        	return err
        }
        //log.Printf("heart beat recv:%s",heartbeat.GetWorknode())
        worknode = heartbeat.GetWorknode()
        s.updateWorker(worknode)
        hr :=pb.HeartbeatResponse{
        	   
        	   Leader:s.masternode,
			}
        if err := stream.Send(&hr); err != nil {
        	log.Println("send error",err)
        	s.deleteWorker(worknode)
			return err
		} 

     }

}
func (s *Server) updateWorker(workernode string) {
    _,exists:= s.workers[workernode]
    if exists{
    	return 
    }
    log.Printf("add worker node:%s",workernode)
    s.workers[workernode]=true
}
func (s *Server) deleteWorker(workernode string) {
	log.Printf("delete worker node:%s",workernode)
	delete(s.workers,workernode)
}
func (s *Server) workerHandler(w http.ResponseWriter, r *http.Request) {
	
	var workers []string
	
	for wk,_:=range s.workers{
		workers=append(workers,wk)
	}
	//workers=append(workers,"123")
    
    writeJsonQuiet(w, r, http.StatusOK,workers)

}
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	ret := ClusterStatusResult{
		IsLeader: s.raftServer.Leader() == s.raftServer.Name(),
		Peers:    s.Peers(),
		Leader:   s.raftServer.Leader(),
	}
	writeJsonQuiet(w, r, http.StatusOK, ret)
}
func writeJson(w http.ResponseWriter, r *http.Request, httpStatus int, obj interface{}) (err error) {
	var bytes []byte
	if r.FormValue("pretty") != "" {
		bytes, err = json.MarshalIndent(obj, "", "  ")
	} else {
		bytes, err = json.Marshal(obj)
	}
	if err != nil {
		return
	}
	callback := r.FormValue("callback")
	if callback == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		_, err = w.Write(bytes)
	} else {
		w.Header().Set("Content-Type", "application/javascript")
		w.WriteHeader(httpStatus)
		if _, err = w.Write([]uint8(callback)); err != nil {
			return
		}
		if _, err = w.Write([]uint8("(")); err != nil {
			return
		}
		fmt.Fprint(w, string(bytes))
		if _, err = w.Write([]uint8(")")); err != nil {
			return
		}
	}

	return
}

// wrapper for writeJson - just logs errors
func writeJsonQuiet(w http.ResponseWriter, r *http.Request, httpStatus int, obj interface{}) {
	if err := writeJson(w, r, httpStatus, obj); err != nil {
		//log.Infof("error writing JSON %s: %v", obj, err)
	}
}
/*
func (s *Server) readHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	value := s.db.Get(vars["key"])
	w.Write([]byte(value))
}

func (s *Server) writeHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	// Read the value from the POST body.
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	value := string(b)

	// Execute the command against the Raft server.
	_, err = s.raftServer.Do(command.NewWriteCommand(vars["key"], value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}*/
