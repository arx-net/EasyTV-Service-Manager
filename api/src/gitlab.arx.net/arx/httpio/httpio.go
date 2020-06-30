package httpio

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func ReadJSON(r *http.Request) (map[string]interface{}, error) {
	// Read all
	json_str, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}

	// Unmarshal
	var data map[string]interface{}
	err = json.Unmarshal(json_str, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func WriteJSON(w http.ResponseWriter, code int, data map[string]interface{}) {
	json_string, ok := json.Marshal(data)

	if ok != nil {
		log.Fatal(ok)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(json_string)
}

func WriteText(w http.ResponseWriter, code int, data string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	w.Write([]byte(data))
}

func WriteHTML(w http.ResponseWriter, code int, data string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(code)
	w.Write([]byte(data))

}
