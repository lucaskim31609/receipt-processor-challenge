# Receipt Processor Challenge (Lucas Kim)

## Overview

This project is my implementation of the Fetch Receipt Prcoessor Challenge. It exposes an HTTP API for submitting receipts and retrieving the calculated points.

This service was built using Go and relies only on standard libraries and the `github.com/google/uuid` package for ID generation.

## Functionality

The core purpose of this service is to calculate points for receipts according to specific rules. It provides two main API endpoints:

1.  **`POST /receipts/process`**
    * Accepts a JSON payload representing a receipt (see `examples/` directory or `api.yml` for structure).
    * Validates the incoming receipt data against the API specification.
    * Calculates points based on the rules outlined in the challenge description.
    * Stores the calculated points associated with a newly generated unique receipt ID.
    * Returns a JSON response containing the unique ID, e.g., `{ "id": "..." }`.

2.  **`GET /receipts/{id}/points`**
    * Accepts a receipt ID as part of the URL path.
    * Looks up the points previously calculated and stored for that ID.
    * Returns a JSON response containing the point total, e.g., `{ "points": 109 }`.

**Important Note:** As per the requirements, data persistence is **not** implemented. All receipt IDs and their associated points are stored **in memory** and will be lost when the application stops or restarts.

## File Structure

* `main.go`: Contains the main application setup, HTTP server configuration, and request routing. HTTP handlers are also defined here.
* `receipt.go`: Defines the data structures (`Receipt`, `Item`, etc.) and contains the core logic for validating receipts and calculating points.
* `helpers.go`: Contains small utility functions (e.g., for sending JSON responses).
* `api.yml`: The OpenAPI 3.0 specification defining the API contract.
* `go.mod`, `go.sum`: Go module files defining dependencies.
* `examples/`: Contains sample JSON files (`morning-receipt.json` & `simple-receipt.json`) that can be used for testing the `/receipts/process` endpoint.
* `.gitignore`: Specifies intentionally untracked files for Git.
* `README.md`: This file.

## Running the Application

**Instructions:**

1.  **Clone the Repository** (or ensure you have the code):
    ```bash
    # git clone ... (if applicable)
    ```
2.  **Navigate to the Directory:**
    ```bash
    cd path/to/receipt-processor-challenge
    ```
3.  **Ensure Dependencies are Downloaded** (usually handled automatically, but good practice):
    ```bash
    go mod tidy
    ```
4.  **Run the Server:**
    ```bash
    go run .
    ```
    * The server will start, and you should see log output indicating it's listening, typically on port 8080.
    * `{"time":"...","level":"INFO","msg":"Server starting...","port":"8080"}`
    * You can specify a different port by setting the `PORT` environment variable: `PORT=8081 go run .`

## Using the API (Examples)

You can use tools like `curl` to interact with the running service. Make sure the server is running first.

* **Process a Receipt:**
    ```bash
    # Run from the project root directory
    curl -X POST http://localhost:8080/receipts/process \
         -H "Content-Type: application/json" \
         --data "@examples/simple-receipt.json"
    ```
    * This will return a JSON response like: `{"id":"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"}`. Copy the returned ID.

* **Get Points:**
    ```bash
    # Replace YOUR_RECEIPT_ID with the actual ID you received above
    curl http://localhost:8080/receipts/YOUR_RECEIPT_ID/points
    ```
    * This will return a JSON response like: `{"points":31}` (for the `simple-receipt.json` example).

## API Specification

The formal API contract is defined in the `api.yml` file using the OpenAPI 3.0 standard.

---

Feel free to explore the code! The main logic for validation and points calculation resides in `receipt.go`.

