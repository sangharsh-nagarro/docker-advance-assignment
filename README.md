# User Management Service

This project is a simple user management service built with Go, utilizing Redis for caching and MongoDB for persistent storage. It provides APIs for creating and retrieving user information.

## Project Structure

- `main.go`: Contains the main application logic, including API handlers and database connections.
- `Dockerfile`: Defines the container image for the web service.
- `docker-compose.yml`: Orchestrates the web service, Redis, and MongoDB containers.

## Prerequisites

- Docker
- Docker Compose

## Setup and Running

1. Clone the repository:
   ```
   git clone https://github.com/sangharsh-nagarro/docker-advance-assignment.git
   cd docker-advance-assignment
   ```

2. Create a `.env` file in the project root with the following variables:
   ```
   REDIS_PASSWORD=<your-redis-password>
   MONGO_INITDB_ROOT_USERNAME=<your-mongo-username>
   MONGO_INITDB_ROOT_PASSWORD=<your-mongo-password>
   MONGO_INITDB_DATABASE=development
   ```

3. Build and start the services:
   ```
   docker-compose up --build
   ```

4. The web service will be available at `http://localhost:8080`

## API Documentation

### Health Check

- **URL**: `/health`
- **Method**: `GET`
- **Description**: Check if the service is running.
- **Success Response**:
  - **Code**: 200
  - **Content**: `<h1>healthy</h1>`

### Create User

- **URL**: `/api/user/create`
- **Method**: `POST`
- **Description**: Create a new user.
- **Request Body**:
  ```json
  {
    "name": "John Doe",
    "email": "john@example.com"
  }
  ```
- **Success Response**:
  - **Code**: 201
  - **Content**:
    ```json
    {
      "message": "User created successfully",
      "id": "<inserted-id>"
    }
    ```
- **Error Response**:
  - **Code**: 400
  - **Content**: `Invalid request body`
  - **Code**: 500
  - **Content**: `Failed to create user`

### Get User

- **URL**: `/api/user?email=<user-email>`
- **Method**: `GET`
- **Description**: Retrieve user information by email.
- **URL Params**: 
  - `email=[string]` (required)
- **Success Response**:
  - **Code**: 200
  - **Content**:
    ```json
    {
      "User": {
        "name": "John Doe",
        "email": "john@example.com"
      },
      "DataSource": "cache" // or "database"
    }
    ```
- **Error Response**:
  - **Code**: 400
  - **Content**: `Email parameter is required`
  - **Code**: 404
  - **Content**: `User not found`
  - **Code**: 500
  - **Content**: `Failed to fetch user`

## Development

The project uses Docker Compose's `develop` feature for hot-reloading during development. Any changes to the source code will trigger a rebuild of the web service container.

## Networks

The project uses two Docker networks:
- `frontend`: Exposes the web service to the host machine.
- `private`: An internal network for communication between services.

## Volumes

- `redis-data`: Persistent storage for Redis data.
- `mongo-data`: Persistent storage for MongoDB data.

## Health Checks

All services (Redis, MongoDB, and the web service) have health checks configured to ensure they are running correctly before being considered available.

## CI/CD Pipeline

This project uses GitHub Actions for continuous integration and continuous deployment (CI/CD). The pipeline is defined in `.github/workflows/docker-publish.yml`.

### Workflow Details

- **Trigger**: The workflow is triggered on push to the `main` branch.
- **Environment**: Runs on the latest Ubuntu runner.

### Steps

1. **Checkout**: Fetches the repository code.
2. **Set up Docker Buildx**: Prepares for multi-platform builds.
3. **Login to Docker Hub**: Authenticates with Docker Hub using secrets.
4. **Docker meta**: Generates metadata for Docker tags and labels.
5. **Build and push**: Builds the Docker image and pushes it to Docker Hub.

### Docker Image

- **Repository**: `sangharshseth/docker-advance-assignment`
- **Tags**: `latest`
- **Platforms**: `linux/amd64`

### Secrets

The workflow uses the following secrets:
- `DOCKERHUB_USERNAME`: Your Docker Hub username
- `DOCKERHUB_TOKEN`: Your Docker Hub access token

To use this CI/CD pipeline:
1. Fork this repository.
2. Set up the required secrets in your GitHub repository settings.
3. Push changes to the `main` branch to trigger the workflow.

The pipeline will automatically build and push a new Docker image to Docker Hub whenever changes are pushed to the `main` branch.

## CD Workflow (deploy-local.yml)

This workflow handles the Continuous Deployment (CD) part, deploying the application to your local machine.

#### Workflow Details

- **Trigger**: Runs after the "ci" workflow completes successfully.
- **Environment**: Runs on a self-hosted runner (your local machine).

#### Steps

1. **Checkout**: Fetches the latest code from the repository.
2. **Pull latest images**: Pulls the latest Docker images.
3. **Deploy**: Runs `docker-compose up -d` to deploy the application.
4. **Clean up**: Removes old, unused Docker images.

#### Setup for Local Deployment

To use the CD pipeline for local deployment:

1. Set up a self-hosted runner on your local machine:
   - Go to your GitHub repository settings.
   - Navigate to "Actions" > "Runners".
   - Click "New self-hosted runner" and follow the instructions to set it up on your local machine.
2. Ensure Docker and Docker Compose are installed on your local machine.
3. Make sure your .env file is present in the project root directory with all necessary environment variables.
4. Ensure your local machine has permissions to pull from your Docker Hub repository.

#### Environment Variables

The CD workflow uses the .env file located in your project directory on your local machine. Since the workflow runs on your local machine as a self-hosted runner, it has direct access to this file. But in a real production environment we should use some kind of secrets manager
