# Trading Engine

## Introduction

Welcome to the Trading Engine project, a practical and efficient simulation of a stock trading platform. This engine, crafted in Go, is tailored for efficient processing of concurrent buy and sell stock orders. It utilizes Redis for streamlined order management. This project is particularly focused on demonstrating the effective handling of high-frequency trading operations, showcasing both simplicity and efficiency in its design and functionality.

## Technologies Used

This project is built using the following technologies and tools:

- **Go (Golang)**: Primary programming language, chosen for its performance and concurrency support. Preference given to official Go libraries to ensure optimal performance.
- **PostgreSQL**: Robust relational database used for storing and querying persistent data, ensuring data integrity and efficient access.
- **Redis**: Utilizes sorted sets to manage order prices as keys for queue access, and implements FIFO (First-In-First-Out) queues for efficient order processing. 
- **Docker**: Used for containerization, ensuring consistent environments and ease of deployment.

## Installation and Setup

This project uses Docker for Redis and PostgreSQL, and requires the Go-Migrate CLI tool. Follow these steps to get started:

1. **Clone the Repository**:
    ```
   git clone https://github.com/MaxwellKuo47/tradingEngine.git
   cd tradingEngine
   ```

2. **Install Go-Migrate CLI**:
- Follow the instructions at [Go-Migrate-CLI-doc](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) to install the CLI tool.

3. **Setup Docker Containers**:
- Pull Docker images:
  ```
  make docker/pull/postgres
  make docker/pull/redis
  ```
- Run PostgreSQL and Redis containers:
  ```
  make docker/create/container/postgresql
  make docker/create/container/redis
  ```

4. **Database Setup**:
- Create the database and user:
  ```
  make docker/create/db
  make docker/create/dbExtension
  make docker/create/db/user
  ```
- Apply database migrations.
  ```
  make db/migrate/up
  ```

5. **Build the Application and Run**:
- Build and run the application
  ```
  make build/api
  ./bin/linux_amd64/api   # for linux os

  # Or directly run it
  go run ./cmd/api
  ```
## Database Schema
![image](https://github.com/MaxwellKuo47/tradingEngine/blob/main/assets/db/schema.png)
The Trading Engine uses a PostgreSQL database with the following key tables:
- `users`: Stores user information including name, email, and hashed password.
- `tokens`: Manages authentication tokens and their expiry.
- `orders`: Records details of buy and sell orders, including quantity, price, and status.
- `trades`: Records consumed buy/sell order and their executed time.
- `user_stock_balances`: Tracks users' stock quantities.
- `user_wallets`: Maintains users' wallet balances.
- `stocks`: Lists available stocks in the trading platform.

Indexes and foreign keys are used for optimized query performance and data integrity. The schema is designed to support efficient order processing and user management in a high-frequency trading environment.

## Order Processing Mechanism

### Overview
Our trading engine handles incoming orders by integrating postgreSQL with Redis to ensure efficient order management and execution. Here's how the process works:

### Order Recording
- Upon receiving an order, its details are recorded in the postgreSQL for persistence and tracking.

### Order Queuing and Prioritization
![image](https://github.com/MaxwellKuo47/tradingEngine/blob/main/assets/db/DS.png)
- Buy/Sell orders are managed using Redis data structures:
  - **Sorted Sets (Heaps)**: Each order's price is added to a Redis sorted set, serving as a heap. Buy orders are organized in descending order, while sell orders are in ascending order.
  - **FIFO Queues**: Corresponding to each price point in the heap, there's a Redis list (queue) that stores orders at that price. Orders in the same queue have identical prices, differing only in their arrival times.

### Order Matching and Execution
- Each `stock_id` has buy and sell consumers continuously get current stock price and monitor the heaps.
- The buy consumer examines the highest value in the buy heap, and the sell consumer looks at the lowest in the sell heap.
- If an order meets the criteria for execution (i.e., the current stock price 
less than or equal to the highest buy price in the heap), the order is consumed.
- After the execution, the trade details and user balances are updated in the database. Additionally, if a queue becomes empty, the corresponding price in the heap is also removed.

## API Documentation

### User Registration
Register a new user with their name, email, and password.
- **Method:** `POST`
- **Path:** `http://localhost:8080/v1/users`
- **Example Input:**
  ```json
  {
      "name": "Test001",
      "email": "test001@example.com",
      "password": "pa$$w0rd12345"
  }
  ```
- **Example Output:**
    ```json
    {
        "user": {
            "id": 1,
            "created_at": "2023-12-18T13:20:54.495198Z",
            "name": "Maxwell",
            "email": "max8783890@gmail.com"
    }
    ```

### User Login
Authenticate a user and get an authentication token to use with other endpoints that require authorization.
- **Method:** `POST`
- **Path:** `http://localhost:8080/v1/users/authentication`
- **Example Input:**
  ```json
  {
    "email": "test001@example.com",
    "password": "pa$$w0rd12345"
  }
  ```
- **Example Output:**
    ```json
    {
        "authentication_token": {
            "token": "FCK3IW3D4IAZWUIAVVDHEGHXAY",
            "expiry": "2023-12-19T21:20:58.206274+08:00"
        }
    }
    ```
### Create Order
Place a new order for buying or selling stocks. This endpoint requires a valid authentication token.

- **Method:** `POST`
- **Path:** `http://localhost:8080/v1/orders`
- **Required Header:** `Authorization: Bearer <token>`
- **Example Input:**
  ```json
    {
        "stock_id": 1,
        "type": 0, // 0: buy, 1: sell
        "quantity": 1,
        "price_type": 1, // 0: market, 1: limit
        "price": 90
    }
    ```

- **Example Output:**
    ```json
    {
        "message": "order create successfully"
    }
    ```
### Stock Price Adjust
Use for simulating stock price change to trigger buy/sell order consuming

- **Method:** `POST`
- **Path:** `http://localhost:8080/v1/stockValueChangeHandler`
- **Example Input:**
  ```json
    {
        "stock_id": 1, // currently only 1 - 5
        "price": 90 // every stock default price is 100
    }
    ```

- **Example Output:**
    ```json
    {
        "params": {
            "stock_id": 1,
            "price": 90
        }
    }
    ```
## Future Enhancements

Given more time and resources, the following improvements are planned for the Trading Engine:

1. **Load Testing with K6**: Implement K6 scripts for more comprehensive load testing. This will automate the testing process, eliminating the need for manual testing with tools like Postman or curl.

2. **Unit Testing**: Develop a suite of unit tests to ensure code robustness and reliability, further enhancing the quality and stability of the application.
