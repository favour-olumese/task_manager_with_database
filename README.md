# Implementing Authentication and Authorization with JWT for Task Management API (A2SV Task)
A task manager built with the Gin framework using MongoDB for its database.

## Objective
The objective of this task is to refactor the existing Task Management API codebase using Clean Architecture principles. This refactor aims to improve the maintainability, testability, and scalability of the application by organizing the code into distinct layers with clear separation of concerns.

## Requirements:
1. Analyze the existing Task Management API codebase and identify areas for improvement based on Clean Architecture principles.
2. Refactor the codebase into separate layers representing different architectural components, such as domain, use cases, and infrastructure.
3. Define clear boundaries between layers to enforce dependencies from higher-level layers to lower-level layers.
4. Implement domain models representing core business entities and logic, ensuring they remain independent of external frameworks or libraries.
5. Define use cases as the central component of the application, encapsulating the application's business logic and orchestrating interactions between different layers.
6. Implement interfaces to abstract external dependencies, such as data access mechanisms, allowing for easy substitution and testing.
7. Organize the codebase into distinct packages or modules representing different architectural layers (e.g., domain, usecase, repository, delivery).
8. Update the API endpoints to interact with the use cases layer, ensuring that business logic is centralized and reusable.

  
## Instructions
1. Conduct a thorough review of the existing Task Management API codebase to understand its structure and functionality.
2. Identify areas where the codebase could benefit from restructuring based on Clean Architecture principles, such as separation of concerns and dependency inversion.
3. Refactor the codebase into separate layers, with clear boundaries and dependencies between layers.
4. Implement domain models representing core business entities and logic, ensuring they are decoupled from external frameworks or libraries.
5. Define use cases to encapsulate the application's business logic, orchestrating interactions between different layers and enforcing business rules.
6. Implement interfaces to abstract external dependencies, such as data access mechanisms, allowing for easy substitution and testing.
7. Organize the codebase into packages or modules representing different architectural layers, with clear naming conventions to indicate their purpose.
8. Update the API endpoints to interact with the use cases layer, ensuring that business logic is centralized and reusable across different delivery mechanisms.

## Folder Structure:
Follow the following folder structure for this task

task_manager/
├── Delivery/
│   ├── main.go
│   ├── controllers/
│   │   └── controller.go
│   └── routers/
│       └── router.go
├── Domain/
│   └── domain.go
├── Infrastructure/
│   ├── auth_middleWare.go
│   ├── jwt_service.go
│   └── password_service.go
├── Repositories/
│   ├── task_repository.go
│   └── user_repository.go
└── Usecases/
│   ├── task_usecases.go
│   └── user_usecases.go
├── docs/
│   └── api_documentation.md
└── go.mod
  

* Delivery/: Contains files related to the delivery layer, handling incoming requests and responses.
    * main.go: Sets up the HTTP server, initializes dependencies, and defines the routing configuration.
    * controllers/controllers.go: Handles incoming HTTP requests and invokes the appropriate use case methods.
    * routers/routers.go: Sets up the routes and initializes the Gin router.
* Domain/: Defines the core business entities and logic.
    * domain.go: Contains the core business entities such as Task and User structs. 
* Infrastructure/: Implements external dependencies and services.
    * auth_middleWare.go: Middleware to handle authentication and authorization using JWT tokens.
    * jwt_service.go: Functions to generate and validate JWT tokens.
    * password_service.go: Functions for hashing and comparing passwords to ensure secure storage of user credentials.
* Repositories/: Abstracts the data access logic.
    * task_repository.go: Interface and implementation for task data access operations.
    * user_repository.go: Interface and implementation for user data access operations.
* Usecases/: Contains the application-specific business rules.
    * task_usecases.go: Implements the use cases related to tasks, such as creating, updating, retrieving, and deleting tasks.
    * user_usecases.go: Implements the use cases related to users, such as registering, logging in.

## Evaluation Criteria:
* Adherence to Clean Architecture principles, including separation of concerns, dependency inversion, and clear boundaries between layers.
* Clear organization of the codebase into distinct layers with well-defined responsibilities.
* Implementation of domain models and use cases representing core business logic, ensuring they remain decoupled and reusable.
* Proper abstraction of external dependencies using interfaces, allowing for easy substitution and testing.
* Maintenance of backward compatibility with the existing API functionality through careful refactoring.
* Thorough documentation of the refactored architecture, including design decisions and guidelines for future development.
* Completion of unit tests for critical components to ensure correctness and maintainability.
Compliance with the provided instructions and requirements

## Note
* Clean Architecture provides a flexible and scalable approach to designing software systems. Focus on achieving a clear separation of concerns and organizing the codebase into layers that facilitate maintainability and testability.
