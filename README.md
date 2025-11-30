# Background

**This project supports simultaneous multi-currency conversion**, displaying exchange rates for multiple currencies at once.

**Demo**: https://ex.luchang.xyz

Unlike most currency conversion tools that only support 1-to-1 conversion (requiring users to repeatedly switch settings for different currency pairs), this dramatically improves efficiency for users dealing with international transactions or multi-currency portfolios.

The application fetches live exchange rate data from the [Frankfurter API](https://github.com/lineofflight/frankfurter) project and implements short-term caching to optimize performance.

## Tech Stack

### Backend (Server)

- **Language**: Go 1.25
- **HTTP Server**: Standard `net/http` package
- **Caching**: In-memory cache with TTL
- **Container**: Multi-stage Docker build with distroless base image

### Frontend (Client)

- **Framework**: Svelte 5
- **Build Tool**: Vite 7
- **Package Manager**: npm
- **Web Server**: Nginx (for production)

### Infrastructure

- **Containerization**: Docker
- **CI/CD**: GitHub Actions
- **Deployment**: AWS EC2
- **Registry**: Docker Hub

## Project Structure

```
lc-exchanger/
├── client/           # Frontend Svelte application
│   ├── src/
│   ├── public/
│   └── Dockerfile
├── server/           # Backend Go API service
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
└── .github/
    └── workflows/    # CI/CD pipelines
```

## Development

### Prerequisites

- Node.js 22+
- Go 1.25+
- Docker

### Local Setup

**Frontend (development mode):\*\***

```bash
cd client
npm install
npm run dev
```

**Backend (development mode):**

```bash
cd server
go run main.go
```

## Deployment

The application uses automated CI/CD via GitHub Actions:

- Frontend builds are deployed to `/var/www/exchanger` on EC2
- Backend Docker images are built, pushed to Docker Hub, and deployed to EC2
- Deployments trigger automatically on pushes to the `main` branch

## License

MIT
