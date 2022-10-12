# Future Backend API - Vincent Gonzalez

## Runtime Environment
The minimum version of Go that is required to run this solution is `1.18`. This version is also defined in the `go.mod` file of this solution. However, I have compiled and run this solution using version `1.19.2` of Go.

## How to run
A makefile has been included in the base directory. Use `make buildRun` in a terminal in the base directory to build and execute the program. Use `make build` to only compile the program. Use `make run` to run the program after it has already been compiled.

## Dependencies
This program uses the following 3rd party libraries:
- [HttpRouter](https://github.com/julienschmidt/httprouter)
```
go get github.com/julienschmidt/httprouter
```
- [go-sqlite3](https://github.com/mattn/go-sqlite3)
```
go get github.com/mattn/go-sqlite3
```

You may need to ``` go get ``` these packages if the program fails to compile.

## A note on date conversions
There are many instances in this program where date conversions happen. This is due to the lack of a date type in sqlite and would not be needed as much in a DBMS that supports date types such as PostgreSQL or MySQL.

## Return object structure
### On success
```
{
    "status": "success",
    "data": "the data requested as {} or []"
}
```
### On failure
```
{
    "status": "fail",
    "data": {
        "prop_name": "a message attached to a property that failed"
    }
}
```
### On error
```
{
    "status": "error",
    "message": "a message describing the error"
}
```

## Endpoints
### Get a trainer's available appointments between 2 dates
```
GET http://localhost:8081/appointments?trainer_id=1&starts_at=2019-01-24T07:00:00-08:00&ends_at=2019-01-24T18:00:00-08:00
```
Returns
```
{
    "status": "success",
    "data": [
        {
            "id": 11,
            "trainer_id": 1,
            "user_id": 1,
            "starts_at": "2019-01-24T07:00:00-08:00",
            "ends_at": "2019-01-24T18:00:00-08:00"
        },
        {
            "id": 12,
            "trainer_id": 2,
            "user_id": 50,
            "starts_at": "2019-01-24T07:00:00-08:00",
            "ends_at": "2019-01-24T18:00:00-08:00"
        }
    ],
}
```

### Post a new appointment
```
POST http://localhost:8081/appointments

request body:
{
    "trainer_id": 1,
    "user_id": 1,
    "starts_at": "2019-01-24T07:00:00-08:00",
    "ends_at": "2019-01-24T18:00:00-08:00"
}
```
Returns
```
{
    "status": "success",
    "data": {
        "id": 11,
        "trainer_id": 1,
        "user_id": 1,
        "starts_at": "2019-01-24T07:00:00-08:00",
        "ends_at": "2019-01-24T18:00:00-08:00"
    }
}
```

### Get a trainer's scheduled appointments
:id is the trainer's trainer_id
```
http://localhost:8081/trainers/:id/appointments
```
Returns
```
{
    "status": "success",
    "data": [
        {
            "id": 11,
            "trainer_id": 1,
            "user_id": 1,
            "starts_at": "2019-01-24T07:00:00-08:00",
            "ends_at": "2019-01-24T18:00:00-08:00"
        }
    ]
}
```

### **Makefile**
Contains the CLI commands for the project. It also contains some environment variables that are used within the CLI commands.

The file includes the following commands:
- `build` - changes to the source directory, compiles the program, and places the executable into a build directory.
- `buildRun` - combines the `build` and `run` commands into one.
- `run` - executes the compiled program.
