# Chat App

This is a simple real-time chat application built using Golang, React, and MongoDB.

## Design

MongoDB is used for the database. The backend is implemented in Go mostly with the standard library. The frontend is implemented in TypeScript using React.

Local Kubernetes (minikube) is used for deployment. Manifests have been included in `deployment/`.

The backend exposes a REST API that acts as a control plane and a websocket endpoint (`/ws`) to stream messages in real time. This API is fully documented in a separate README in `backend/`.

JWT tokens are used for authentication, and bcrypt is used to hash passwords for storage in the database.

Upvote race conditions are handled using transactions, allowing us to scale the backend in the future if desired.

## To-dos

* [ ] Use session storage to persist session across page refreshes.
* [ ] Implement websocket heartbeat logic on frontend.
* [ ] Add logout functionality.
* [ ] Un-ugly the frontend.
* [ ] Testing.
* [x] Add authentication to `/ws` endpoint.
* [ ] Stream vote updates through websocket to update vote counts dynamically.

## Running

The application can be run out of the box using minikube.

First, start up minikube with at least two nodes. I found that `minikube start --nodes=2` didn't work for me, so if that is the case for you, start minikube with `minikube start` and then add a node with `minikube node add`.

Next, run `minikube tunnel` so that the frontend and backend have visible IP addresses we can connect to.

Next, apply all manifests with `kubectl apply -Rf deployment/`, which will recursively apply manifests in `deployment/`. The Docker images for the deployments are stored on Docker hub, so there is no need to build anything here.

At this point, you should be able to connect to the frontend at `localhost:3000`.

When you are done testing, delete the resources with `kubectl delete -Rf deployment/`.
