package main
 
import (
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"net/http"
	"strings"
	"time"
)
 
// Representation of the metadata an asset should have.
type Asset struct {
	Fid     string
	Created time.Time
	Origin  string
	Ma01    bool
	Tx01    bool
}
 
func get_bad_file(session *gocql.Session, dc string, w http.ResponseWriter, r *http.Request) {
	/*
		Queries Cassandra for a file that is not in the specified datacenter.
		Presumably, a worker will get this file and add it locally.
	*/
 
	// Instantiate the struct.
	file := Asset{}
 
	// Compose the CQL query.
	query := fmt.Sprintf(`
		SELECT fid, created, origin, ma01, tx01
		FROM file_ledger
		WHERE %s = false
		LIMIT 1`, dc)
 
	// Execute the query and store the results in the struct above.
	q := session.Query(query)
	q.Scan(&file.Fid, &file.Created, &file.Origin, &file.Ma01, &file.Tx01)
 
	log.Printf("%s requested bad file for %s: %s, Originally created at %s in %s. Exists in MA01: %t. Exists in TX01: %t.", strings.Split(r.RemoteAddr, ":")[0], dc, file.Fid, file.Created, strings.ToUpper(file.Origin), file.Ma01, file.Tx01)
 
	// Converts the struct to json.
	json, err := json.Marshal(file)
	if err != nil {
		log.Printf("Failed to fetch file info: %s\n", err)
	}
 
	// Checks for empty result.
	if file.Fid != "" {
		// Writes json result payload.
		w.Write(json)
	} else {
		// Returns 'ok' result.
		ok_resp := []byte(`{"all":"good"}`)
		w.Write(ok_resp)
	}
}
 
func add_new_file(session *gocql.Session, w http.ResponseWriter, r *http.Request) {
	/*
		Adds a new file to Cassandra via a POST.
	*/
 
	// Instantiate the struct.
	file := Asset{}
 
	// Decodes the json POST.
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&file)
	if err != nil {
		log.Printf("Failed to decode: %s\n", err)
	}
 
	// We create the creation time instead of the client.
	created := time.Now()
 
	log.Printf("%s added new file: %s, Originally created at %s in %s. Exists in MA01: %t. Exists in TX01: %t.", strings.Split(r.RemoteAddr, ":")[0], file.Fid, file.Created, strings.ToUpper(file.Origin), file.Ma01, file.Tx01)
 
	// Execute the insert, at CL:ALL.
	if err := session.Query(`
		INSERT INTO file_ledger (fid, created, origin, ma01, tx01)
		VALUES (?, ?, ?, ?, ?) 
		IF NOT EXISTS`, &file.Fid, created, &file.Origin, &file.Ma01, &file.Tx01).Consistency(gocql.All).Exec(); err != nil {
		log.Printf("Failed to insert: %s\n", err)
		fail_resp := []byte(`{"insert":"fail"}`)
		w.Write(fail_resp)
	} else {
		ok_resp := []byte(`{"insert":"success"}`)
		w.Write(ok_resp)
	}
}
 
func update_file(session *gocql.Session, dc string, w http.ResponseWriter, r *http.Request) {
	/*
		Updates a new file in Cassandra via a POST.
	*/
 
	// Instantiate the struct.
	file := Asset{}
 
	// Decodes the json POST.
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&file)
	if err != nil {
		log.Printf("Failed to decode: %s\n", err)
	}
 
	log.Printf("%s toggled file %s as true in %s", strings.Split(r.RemoteAddr, ":")[0], file.Fid, strings.ToUpper(dc))
 
	// Compose the CQL query.
	query := fmt.Sprintf(`
		UPDATE file_ledger
		SET %s = true
		WHERE fid = '%s'`, dc, file.Fid)
 
	// Execute the insert, at ConsistancyLevel:ALL.
	if err := session.Query(query).Consistency(gocql.All).Exec(); err != nil {
		log.Printf("Failed to update: %s\n", err)
		fail_resp := []byte(`{"update":"fail"}`)
		w.Write(fail_resp)
	} else {
		ok_resp := []byte(`{"update":"success"}`)
		w.Write(ok_resp)
	}
 
}
 
func main() {
	// Defines the Cassandra Cluster.
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "phylactery"
	cluster.Consistency = gocql.One
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal("Error creating Cassandra session: %v", err)
	}
	defer session.Close()
 
	// Call /file/bad/dc to get a file NOT in that dc but in others.
	// Example:
	//  curl http://localhost:8080/file/bad/tx01
	http.HandleFunc("/file/bad/ma01", func(w http.ResponseWriter, r *http.Request) { get_bad_file(session, "ma01", w, r) })
	http.HandleFunc("/file/bad/tx01", func(w http.ResponseWriter, r *http.Request) { get_bad_file(session, "tx01", w, r) })
 
	// Call /file/add/dc to toggle a file in that dc as present/true.
	// Example:
	//  curl -X POST -d "{\"Fid\":\"1234.fid\"}" http://localhost:8080//file/add/tx01
	http.HandleFunc("/file/add/ma01", func(w http.ResponseWriter, r *http.Request) { update_file(session, "ma01", w, r) })
	http.HandleFunc("/file/add/tx01", func(w http.ResponseWriter, r *http.Request) { update_file(session, "tx01", w, r) })
 
	// Call /file/new to add a new file to the file ledger.
	// Example:
	//   curl -X POST -d "{\"Fid\":\"1234.fid\",\"Origin\":\"ma01\",\"Ma01\":true,\"Tx01\":false}" http://localhost:8080/file/new
	http.HandleFunc("/file/new", func(w http.ResponseWriter, r *http.Request) { add_new_file(session, w, r) })
 
	http.ListenAndServe(":8080", nil)
}
