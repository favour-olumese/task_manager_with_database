# Enhancing Task Management API with Persistent Data Storage using MongoDB and the Mongo Go Driver (A2SV Task)
A task manager built with the Gin framework using MongoDB for its database.

## Objective
The objective of this task is to extend the existing [Task Management API](https://github.com/favour-olumese/task_manager) with persistent data storage using MongoDB and the Mongo Go Driver. This enhancement will replace the in-memory database with MongoDB to provide data persistence across API restarts.

## Requirements:
* Integrate MongoDB as the persistent data storage solution for the Task Management API.
* Use the Mongodb Go Driver to interact with the MongoDB database from the Go application.
* Update the existing API endpoints to perform CRUD operations using MongoDB for data storage.
* Ensure proper error handling and validation of MongoDB operations.
* Validate the correctness of data stored in MongoDB by retrieving and verifying task information.
* Update the documentation to include instructions for configuring and connecting to MongoDB.
* Test the API endpoints to verify that data is stored and retrieved correctly from MongoDB.
* Ensure that the API remains backwards compatible with the previous version, maintaining the same endpoint structure and behaviour.
  
## Instructions
* Set up a MongoDB instance either locally or using a cloud service provider.
* Install the Mongodb Go Driver package (go.mongodb.org/mongo-driver) if not already installed.
* Update the existing API codebase to replace the in-memory database implementation with MongoDB.
* Configure the application to connect to the MongoDB instance using appropriate connection parameters.
* Implement CRUD operations for tasks using the MongoDB Go Driver, including:
    * Creating new tasks
    * Retrieving a list of all tasks
    * Retrieving details of a specific task by ID
    * Updating an existing task
    * Deleting a task
* Ensure that proper error handling is implemented for MongoDB operations, including network errors, database errors, and validation errors.
* Test the API endpoints using Postman or similar tools to verify that tasks can be created, retrieved, updated, and deleted successfully.
* Verify the correctness of data stored in MongoDB by querying the database directly or using MongoDB Compass.
* Document the MongoDB integration process, including any configuration settings or prerequisites required to connect to MongoDB.
* Update the API documentation to reflect changes related to MongoDB integration, including any modifications to request and response formats.

## Folder Structure:
Follow the following folder structure for this task
task_manager/  
├── main.go  
├── controllers/  
│   └── task_controller.go  
├── models/  
│   └── task.go  
├── data/  
│   └── task_service.go  
├── router/  
│   └── router.go  
├── docs/  
│   └── api_documentation.md  
└── go.mod  

* main.go: Entry point of the application.
* controllers/task_controller.go: Handles incoming HTTP requests and invokes the appropriate service methods.
* models/: Defines the data structures used in the application.
* data/task_service.go: Contains business logic and data manipulation functions. Implement the ORM/ODM code here.
* router/router.go: Sets up the routes and initializes the Gin router, and defines the routing configuration for the API.
* docs/api_documentation.md: Contains API documentation and other related documentation.
* go.mod: Defines the module and its dependencies.

## Evaluation Criteria:
* Successful integration of MongoDB as the persistent data storage solution for the Task Management API.
* Correct implementation of CRUD operations using the Mongodb Go Driver.
* Proper error handling for MongoDB operations and network/database errors.
* Verification of data correctness by testing API endpoints and querying MongoDB directly.
* Clarity and completeness of documentation for MongoDB integration and API changes.
* Compliance with the provided instructions and requirements.

## Note
* Ensure that the API remains functional and backwards compatible with the previous version after integrating MongoDB. Test all existing endpoints to verify their continued functionality.
* MongoDB offers various features such as indexes, aggregation pipelines, and transactions. While not required for this task, consider exploring these features for future enhancements to the API.

