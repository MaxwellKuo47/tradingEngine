# Trading Engine

## Introduction

Welcome to the Trading Engine project, a practical and efficient simulation of a stock trading platform. This engine, crafted in Go, is tailored for efficient processing of concurrent buy and sell stock orders. It utilizes Redis for streamlined order management. This project is particularly focused on demonstrating the effective handling of high-frequency trading operations, showcasing both simplicity and efficiency in its design and functionality.

## Technologies Used

This project is built using the following technologies and tools:

- **Go (Golang)**: Primary programming language, chosen for its performance and concurrency support. Preference given to official Go libraries to ensure optimal performance.
- **PostgreSQL**: Robust relational database used for storing and querying persistent data, ensuring data integrity and efficient access.
- **Redis**: Utilizes sorted sets to manage order prices as keys for queue access, and implements FIFO (First-In-First-Out) queues for efficient order processing. 
- **Docker**: Used for containerization, ensuring consistent environments and ease of deployment.

Each technology has been selected to optimize the performance and scalability of the trading engine, ensuring quick processing of high-frequency trading data.

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
- Build the application
  ```
  make build/api
  ```
- Select your system and run
  ```
  # for linux os
  ./bin/linux_amd64/api
  ```