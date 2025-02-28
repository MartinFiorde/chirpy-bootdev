package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func decodeHandler(w http.ResponseWriter, r *http.Request){
    type parameters struct {
        Body string `json:"body"`
    }

    decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err := decoder.Decode(&params)
    
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
				log.Printf("Error marshalling JSON: %s", err)
				w.WriteHeader(500)
				return
		}
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(500)
		w.Write(dat)
		return
    }

	if len(params.Body) > 140 {
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
				log.Printf("Error marshalling JSON: %s", err)
				w.WriteHeader(500)
				return
		}
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(400)
		w.Write(dat)
		return
    }
    
	type returnVals struct {
		Valid bool `json:"valid"`
	}
	respBody := returnVals{
		Valid: true,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(400)
			return
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write(dat)
}
