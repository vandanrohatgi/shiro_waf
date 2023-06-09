package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
)

func IsRequestBlocked(r *http.Request, rule Rules) (bool, error) {

	log.Debug("Triggering Blocking logic...")
	// Check method
	methodDecision, err := checkMethod(r, rule)
	if err != nil || methodDecision {
		return true, err
	}
	log.Debugf("Method Block: %t", methodDecision)
	// Check Body
	bodyDecision, err := checkBody(r, rule)
	if err != nil || bodyDecision {
		return true, err
	}
	log.Debugf("Body Block: %t", bodyDecision)
	//check headers
	headerDecision, err := checkHeaders(r, rule)
	if err != nil || headerDecision {
		return true, err
	}
	log.Debugf("Header Block: %t", headerDecision)
	// request not blocked
	return false, nil
}

// checkMethod takes the incoming request and the associated rule for making blocking decision
func checkMethod(r *http.Request, rule Rules) (bool, error) {
	if !rule.MethodRegex.MatchString(r.Method) {
		return true, fmt.Errorf("request method %s violates defined method %s", r.Method, rule.Method)
	}
	return false, nil
}

// checkBody takes the incoming request and the associated rule for making blocking decision
func checkBody(r *http.Request, rule Rules) (bool, error) {
	body, err := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body)) // Restore request body after reading it
	defer r.Body.Close()

	if err != nil {
		return true, err
	}

	if !rule.BodyRegex.MatchString(string(body)) {
		return true, fmt.Errorf("request body %s violates defined body %s", string(body[:]), rule.Body)
	}

	return false, nil
}

// checkHeaders takes the incoming request and the associated rule for making blocking decision
func checkHeaders(r *http.Request, rule Rules) (bool, error) {
	for key, value := range r.Header {
		valueString := strings.Join(value, ",")
		if !rule.Headers.KeyRegex.MatchString(key) {
			return true, fmt.Errorf("request header %s violates defined header %s",
				key,
				rule.Headers.Key,
			)
		}
		if !rule.Headers.ValueRegex.MatchString(valueString) {
			return true, fmt.Errorf("request header %s violates defined header %s",
				value,
				rule.Headers.Value)
		}
	}
	return false, nil
}

func blockRequest(w *http.ResponseWriter, r *http.Request) {
	_, err := io.Copy(io.Discard, r.Body)
	if err != nil {
		log.Error("Error discarding request body", err)
	} else {
		defer r.Body.Close()
	}
	log.Errorf("Request blocked. No rule found for %s", r.RequestURI)
	http.Error(*w, "Forbidden", http.StatusForbidden)
}
