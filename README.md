# tickets-lambda-get-concerts
A lambda function responsible for returning a list of concerts with ticket information

## Install dependencies
```
make deps
```

## Build
To build the source, targetting AWS Lambda architecture, run
```
make
```

## Deploy
To deploy to AWS, run
```
make deploy
```
This assumes the following:
- AWS CLI is installed and client is authenticated with AWS
- The ARN for the lambda function where the binary is to be deployed is set as an environment variable like so:
```
export ARN=<ARN_GOES_HERE>
```

## Clean
To clean the build directory run
```
make clean
```

## Test
To run all unit tests, run
```
make test
```
N.B. Testing will fail is code coverage is too low. Currently coverage needs to be at least 90%

## Code coverage report
This will produce a code coverage report and serve it using python3's http.server library on localhost port 8000
This assumes python3 is already installed
```
make cover
```

## Go vet
```
make vet
```

## Go fmt
```
make fmt
```
