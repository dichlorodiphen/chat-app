# Chat App

This is a real-time chat application built using Golang, React, and MongoDB. The bulk of it was made in two days for a coding challenge.

## Design

The backend is implemented in Go, and the frontend is implemented in TypeScript using React. MongoDB is used to store credentials and messages.

Local Kubernetes (minikube) is used for deployment. Manifests have been included in `deployment/`.

The backend exposes a REST API that acts as a control plane and a websocket endpoint (`/ws`) to stream messages in real time. This API is fully documented in a separate [README](server/README.md) in `server/`.

JWTs are used for authentication, and bcrypt is used to hash passwords for storage in the database.

Upvote race conditions are handled using transactions, allowing us to scale the backend in the future if desired.

## To-dos

* [x] Persist sessions across page refreshes.
* [x] Add logout functionality.
* [x] Add authentication to `/ws` endpoint.
* [x] Fix websocket heartbeat logic.
* [ ] Stream vote updates through websocket to update vote counts dynamically.
* [ ] Add error messages on frontend for failed authentication.
* [ ] Un-ugly the frontend.
* [ ] Testing.

## Running

The application can be run out of the box using minikube.

First, start up minikube with at least one worker node: `minikube start --nodes 2` (assuming default CPU and memory allocation for worker nodes).

Next, run `minikube tunnel` so that the frontend and backend have visible IP addresses we can connect to.

Next, apply all manifests with `kubectl apply -Rf deployment/`, which will recursively apply manifests in `deployment/`. The Docker images for the deployments are stored on Docker hub, so there is no need to build anything here.

At this point, you should be able to connect to the frontend at `localhost:3000`. The server runs on port `8000`.

When you are done testing, delete the resources with `kubectl delete -Rf deployment/`.
