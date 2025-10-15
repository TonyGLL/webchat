# WebChat

WebChat is a modern, real-time chat application featuring a backend built with Go and a frontend powered by Next.js. The application is designed to be scalable, maintainable, and performant, following industry best practices.

## Features

*   **Real-time Messaging:** Instantaneous message delivery using WebSockets.
*   **User Authentication:** Secure user registration and login with JWT-based authentication.
*   **Multi-room Chat:** Ability to create and join different chat rooms.
*   **Clean Architecture:** A well-structured backend that separates concerns for better maintainability.
*   **Modern Frontend:** A responsive and user-friendly interface built with Next.js and Tailwind CSS.

## Tech Stack

### Backend

*   **Language:** Go
*   **Framework:** Gin
*   **Database:** PostgreSQL
*   **Cache:** Redis
*   **Real-time Communication:** Gorilla WebSocket
*   **Authentication:** JWT (JSON Web Tokens)
*   **ORM:** pgx
*   **Environment Variables:** godotenv
*   **Validation:** go-playground/validator

### Frontend

*   **Framework:** Next.js
*   **Language:** TypeScript
*   **Styling:** Tailwind CSS
*   **State Management:** React Hooks & Context API
*   **Form Handling:** React Hook Form
*   **Schema Validation:** Zod
*   **API Communication:** Axios

## Getting Started

### Prerequisites

*   [Go](https://golang.org/doc/install) (version 1.24.3 or higher)
*   [Node.js](https://nodejs.org/en/download/) (version 20 or higher)
*   [Bun](https://bun.sh/docs/installation)
*   [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
*   [Make](https://www.gnu.org/software/make/)

### Installation and Setup

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/your-username/webchat.git
    cd webchat
    ```

2.  **Backend Setup:**

    *   Navigate to the `backend` directory:
        ```bash
        cd backend
        ```
    *   Create a local environment file from the example:
        ```bash
        cp dev.env.example dev.env
        ```
    *   Update the `dev.env` file with your PostgreSQL and Redis connection details.

3.  **Frontend Setup:**

    *   Navigate to the `frontend` directory:
        ```bash
        cd ../frontend
        ```
    *   Install the dependencies using Bun:
        ```bash
        bun install
        ```

## Usage

You can run the application using Docker (recommended for a production-like environment) or locally for development.

### Running with Docker

This is the simplest way to get the entire application stack (backend, frontend, database, and cache) up and running.

1.  From the root of the project, run:

    ```bash
    docker-compose up --build
    ```

2.  The application will be available at the following URLs:
    *   **Frontend:** `http://localhost:3000`
    *   **Backend:** `http://localhost:8080`

### Running Locally

This method is ideal for development, as it provides hot-reloading for both the backend and frontend.

1.  **Start the database and cache:**

    From the `backend` directory, start the PostgreSQL and Redis containers:

    ```bash
    docker-compose up -d db redis
    ```

2.  **Run the backend:**

    In the `backend` directory, run the following command to start the Go server with hot-reloading:

    ```bash
    make watch
    ```

    The backend will be running on `http://localhost:8080`.

3.  **Run the frontend:**

    In the `frontend` directory, run the following command to start the Next.js development server:

    ```bash
    bun dev
    ```

    The frontend will be running on `http://localhost:3000`.

## Project Structure

The project is organized into two main parts: `backend` and `frontend`.

### Backend

The backend follows the principles of **Clean Architecture**, with a modular structure. Each module (`auth`, `message`, etc.) is self-contained and has the following layers:

*   `domain`: Core business logic and entities.
*   `application`: Use cases and repository interfaces.
*   `persistence`: Implementation of repositories (e.g., PostgreSQL, Redis).
*   `presentation`: API controllers and routes (Gin).

The main entry point for the backend is `cmd/server/main.go`.

### Frontend

The frontend is a standard Next.js application with the following structure:

*   `src/app`: The main application routes and pages.
*   `src/components`: Reusable React components.
*   `src/lib`: Utility functions and shared logic.
*   `src/context`: React context providers.

## API Endpoints

The API is versioned under the `/api/v1` prefix. Here are some of the main endpoints:

*   `GET /health`: Health check endpoint.
*   **Auth:**
    *   `POST /auth/register`: Register a new user.
    *   `POST /auth/login`: Log in a user.
    *   `POST /auth/refresh`: Refresh an access token.
*   **Messages:**
    *   `POST /messages`: Create a new message.
    *   `GET /messages/{roomId}`: Get all messages for a room.
*   **WebSocket:**
    *   `GET /ws`: Upgrade the connection to a WebSocket.

## Contributing

Contributions are welcome! Please feel free to open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.