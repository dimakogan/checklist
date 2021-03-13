package main

import (
  "bytes"
  "math/rand"
  "log"

  "checklist/pir"
)


// Generate a database filled with random bytes
func randomDatabase(nRows int, rowLen int) *pir.StaticDB {
  rows := make([]pir.Row, nRows)
  for i := 0; i < nRows; i++ {
    rows[i] = make(pir.Row, rowLen)
    _, err := rand.Read(rows[i][:])
    if err != nil {
      log.Fatal("rand.Read failed")
    }
  }

  return pir.StaticDBFromRows(rows)
}

func main() {
  // Use the puncturable-set-based PIR scheme
  pirType := pir.DPF

  nRows := 1024
  rowLen := 256
  queryRow := 79

  // Generate a database
  db := randomDatabase(nRows, rowLen)

  // ===== OFFLINE PHASE =====
  //    Client asks for offline hint
  offlineReq := pir.NewHintReq(pirType)

  //    Server responds with hint
  offlineResp, err := offlineReq.Process(*db)
  if err != nil {
    log.Fatal("Offline hint generation failed")
  }

  // XXX Use cryptographic randomness source
  randSource := rand.New(rand.NewSource(27))

  // Initialize the client state
  client := offlineResp.(pir.HintResp).InitClient(randSource)


  // ===== ONLINE PHASE =====
  //    Client generates queries for servers
  queries, recon := client.Query(queryRow)

  //    Servers answer queries
  answers := make([]interface{}, len(queries))
  for i := 0; i < len(queries); i++ {
    answers[i], err = queries[i].Process(*db)
    if err != nil {
      log.Fatal("Error answering query")
    }
  }

  //    Client reconstructs
  row, err := recon(answers)
  if err != nil {
    log.Fatal("Could not reconstruct")
  }

  if !bytes.Equal(row, db.Row(queryRow)) {
    log.Fatal("Incorrect answer returned")
  }
}
