package main

import (
    "encoding/json"
    "fmt"
    "github.com/marni/goigc"
    "log"
    "net/http"
    "regexp"
    "strings"
    "time"
)

//JSON Structs
type IDArray struct {
    Ids []string `json:"ids"`
}
type Metadata struct {
	Uptime 	string 	`json:"uptime"`
	Info    string 	`json:"info"`
	Version string 	`json:"version"`
}

type Track struct {
    H_date time.Time `json:"H_date"`
    Pilot string  `json:"pilot"`
    Glider string `json:"glider"`
    GliderID string `json:"glider_id"`
    TrackLength string `json:"track_length"`
}

type URLRequest struct {
    URL string  `json:"url"`
}

//global variables
var apiStruct Metadata //contains meta information
var start = time.Now() //keeps track of uptime
var LastID int = 0

//arrays
var tracks map[string]igc.Track
var ids IDArray

// HANDLERS

func handlerAPI(w http.ResponseWriter, r *http.Request) {
    //finding uptime
    //I only track uptime until the point of days, as I find it unlikely that this service would
    //be running for weeks on end, let alone months or years.
    elapsedTime := time.Since(start)
    apiStruct.Uptime = fmt.Sprintf("P%dD%dH%dM%dS",
        int(elapsedTime.Hours()/24),    //number of days (no Days method available)
        int(elapsedTime.Hours())%24,    //number of hours
        int(elapsedTime.Minutes())%60,  //number of minutes
        int(elapsedTime.Seconds())%60,  //number of seconds
        )
    json.NewEncoder(w).Encode(apiStruct)

    }

func handlerIGC(w http.ResponseWriter, r *http.Request) {
    if (r.Method=="POST"){
      http.Header.Add(w.Header(), "content-type", "application/json")

        var urlRequest URLRequest

        decoder := json.NewDecoder(r.Body)
        decoder.Decode(&urlRequest)

        var track igc.Track
        var err error
        track, err = igc.ParseLocation(urlRequest.URL)
        if err == nil{
            id := fmt.Sprintf("track id: %d", LastID)
            LastID++
            tracks[id] = track
            json.NewEncoder(w).Encode(id)

        }
    } else if (r.Method=="GET"){
        parts :=strings.Split(r.URL.Path, "/")

        if len(parts)>4{ //Check whether a specific id is being requested
            requestedID := parts[4]
            if requestedID >= string(LastID) && isNumeric(requestedID) {
                track, exists := tracks[requestedID]
                if !exists {
                    //this track does not exist
                    http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
                }

            } else {
                http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
            }
        }else{
            //return array of all ids
        }

    } else { //if
        http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
    }
}


//Utility Functions

func isNumeric(s string) bool { //Checks whether given string is numeric
value, _ := regexp.MatchString("[0-9]+", s)
return value
}

func main() {
apiStruct = Metadata{Uptime: "", Info:"Info for IGC tracks.", Version: "v1" }
	http.HandleFunc("/igcinfo/api/", handlerAPI)
    http.HandleFunc("/igcinfo/api/igc/", handlerIGC)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
