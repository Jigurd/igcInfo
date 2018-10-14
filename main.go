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

//IDArray ... Array keeping track of all IDs.
type IDArray struct {
    Ids []string `json:"ids"`
}
//Metadata ... stores metadata about app
type Metadata struct {
	Uptime 	string 	`json:"uptime"`
	Info    string 	`json:"info"`
	Version string 	`json:"version"`
}

//Track ... stores metadata about track
type Track struct {
    Hdate time.Time `json:"H_date"`
    Pilot string  `json:"pilot"`
    Glider string `json:"glider"`
    GliderID string `json:"glider_id"`
    TrackLength string `json:"track_length"`
}

//URLRequest ... stores URL request
type URLRequest struct {
    URL string  `json:"url"`
}

//global variables
var apiStruct Metadata //contains meta information
var start = time.Now() //keeps track of uptime
//LastID ... keeps track of used IDs
var LastID int

//arrays
var tracks map[string]igc.Track = make(map[string]igc.Track)
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

        } else {
            http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
        }
    } else if (r.Method=="GET"){
        parts :=strings.Split(r.URL.Path, "/")

        if len(parts)>4 { //Check whether a specific id is being requested
            requestedID := parts[4]
            track, exists := tracks[requestedID]
            if !exists {
                //this track does not exist
                http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
            }
            if requestedID >= string(LastID) && isNumeric(requestedID) {
                if len(parts) == 6 {
                    http.Header.Add(w.Header(), "content-type", "application/json")
                    requestedTrack := Track{
                        track.Header.Date,
                        track.Header.Pilot,
                        track.Header.GliderType,
                        track.Header.GliderID,
                        totalDistance(track),
                    }
                    json.NewEncoder(w).Encode(requestedTrack)
                } else if len(parts) == 7 {
                    switch parts[5] {
                    case "pilot":
                        fmt.Fprintf(w, track.Header.Pilot)
                    case "glider":
                        fmt.Fprintf(w, track.Header.GliderType)
                    case "glider_id":
                        fmt.Fprintf(w, track.Header.GliderID)
                    case "track_length":
                        fmt.Fprintf(w, "%f", track.Task.Distance())
                    case "H_date":
                        fmt.Fprintf(w, "%v", track.Header.Date)
                    default:
                        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
                }

            } else {
                http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
            }
        }
        }else{
            //return array of all ids
            http.Header.Add(w.Header(), "content-type", "application/json")
            json.NewEncoder(w).Encode(ids)
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

func totalDistance(t igc.Track) string {
    track := t
    totalDistance := 0.0
    for i := 0; i < len(track.Points)-1; i++ {
        totalDistance += track.Points[i].Distance(track.Points[i+1])
    }

    return fmt.Sprintf("%f", totalDistance)
}

func main() {
    LastID = 0
    ids = IDArray{make([]string, 0)}

    apiStruct = Metadata{Uptime: "", Info:"Info for IGC tracks.", Version: "v1" }
	http.HandleFunc("/igcinfo/api/", handlerAPI)
    http.HandleFunc("/igcinfo/api/igc/", handlerIGC)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
