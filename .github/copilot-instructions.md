This is a Go based repository with with a REST API, Websockets and MQTT for communication. 
It is primarily a framework for building IoT applications. Both as an IoT device as well
as an edge hub. Please follow these guidelines when contributing:

## Code Standards

### Required Before Each Commit
- Run `make fmt` before committing any changes to ensure proper code formatting
- This will run gofmt on all Go files to maintain consistent style

### Development Flow
- Build: `make build`
- Lint: `make vet`
- Test: `make test`
- Full CI check: `make ci` (includes build, fmt, lint, test)

## Repository Structure
- `cmd/`: Main service entry points and executables.
- `data/`: Temporary data storage and caching.
- `examples/`: Example usage of framework parts.
- `mesh/`: Mesh networking for devices.
- `messanger/`: Pub/Sub styles messaging with MQTT and local (non-broker) messaging.
- `docs/`: Documentation.
- `server/`: HTTP UI, REST and Websockets.
- `station/`: Station Manager
- `testing/`: System and integration tests
- `utils/`: Utilities like logging random and timers

## Key Guidelines
1. Follow Go best practices and idiomatic patterns
2. Maintain existing code structure and organization
3. Use dependency injection patterns where appropriate
4. Write unit tests for new functionality. Use table-driven unit tests when possible.
5. Document public APIs and complex logic. Suggest changes to the `docs/` folder when appropriate
