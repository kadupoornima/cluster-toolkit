// Copyright 2026 "Google LLC"
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func main() {
	fmt.Println("--- DEBUG: ATTEMPTING TO START SERVER ---")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dump, _ := httputil.DumpRequest(r, true)
		fmt.Printf("\n[NEW REQUEST]\n%s\n", string(dump))
		fmt.Fprint(w, "Echoing back: Request received!")
	})

	// We'll use 8888 to avoid common port conflicts
	port := ":8888"
	fmt.Printf("Success! Listening on http://127.0.0.1%s\n", port)
	fmt.Println("Keep this terminal window OPEN.")

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Printf("FATAL ERROR: %v\n", err)
	}
}
