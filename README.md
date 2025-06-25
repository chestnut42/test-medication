# How to run

```
# runs go tests (requires docker)
make test

# runs integration tests unsing docker compose
make test-integrate
```

# Solution discussion

## 1. DynamoDB
DynamoDB was chosen because the task seems to fit key-value storage. Advantages of using Dynamo over traditional SQL databases are
the following:
1. It allows for quicker data format evolution. Allows for different objects to hold various fields
2. Point-in-time recovery allows for restoring any data within a backup window. Very handy if we mess up with the data
3. It handles unexpected load spikes without manual interference

## 2. Accepting ID from the client on Create API
If it's backend-to-backend API it's generally a good practice just accept IDs from the caller. Such ID works as deduplication
key as well. That allows for the calling party to strictly establish 1-to-1 relationship between entities in their database
and entities in this service.

Even if the client is a browser, the browser can allow the user to retry creation request without creating a duplicate.

## 3. Owner
Since we are accepting IDs from the caller we need to make sure that different users of the system will never run into 
conflict when using the same IDs.

```
# scenario
UserA creates entity with id = "ID42" -> success
UserB creates different entity with id = "ID42" -> should be a success
# each user from now on sees their own idea of a medication with id = "ID42"
```

If this is a backend-to-backend system, the service must employ some sort of API keys. In this case `Owner` will be some
sort of project ID.

If this is a user facing API the service must have some user authorisation mechanism (here in the service or in the gateway/
load balancer). UserID then will be the `Owner`.

It of course can work in both modes simultaneously. It's up to transport layer to detect the `Owner`, authorise it and pass it down.

## 4. Framework

I've chosen solutions and libraries based on my own experience. Of course that is **important** for the service to be
similar to other services in the code base.

**First** thing is study existing code base and employ solutions and styles that **are already there**. Improvements should
be discussed prior introduction.

### httpx.Serve

[httpx.Serve function](internal/utils/httpx/serve.go) wraps http server in a function that is suitable to run inside an error
group. Of course, http server might be just launched in the main

### Holding a logger in a context
[logx/context.go](internal/utils/logx/context.go) provides a couple of functions that help to put logger into context 
then take it from the context when it's needed. I find it very handy as you don't need to shove logger around as dependency.

## URLs

API URLs are `/v1/medication/...`. The same server also serves `/health` and `/metrics` endpoints. That was done with assumption
that API gateway/load balancer will only allow `/v1/medication/...` URLs from outside. Other URLs must be available from the
internal network **only**

## Integration tests

I assume you already have some framework for that and it is to be used. As an example I wrote just a 
[quick bash script](/test/medication/test.sh)
to demonstrate the service works.

# What to be done to consider this APP production-ready (List of TODO)

## 1. Finalise implementation

### Figure out Dosage structure

I deliberately left `Dosage` as a string. Intuitively it should be `struct {amount: int, unit: string_enum}` but my gut tells me
it's more complex than that. First thing I'd have bombarded experts on the topic with tons of questions.

### Implement Update and Delete

Didn't yet implement those. Please [see here](/internal/storage/medication.go)

- Update should take old version and return 409 if it's not equal
- Delete ideally should leave the object in DB and rather just mark it as deleted. Unless we must conform to GDPR-like protocol

### Implement history table

Ideally all versions of all objects are to be saved using Dynamo transaction. That will save **tons** of time once we will
encounter corrupted data situation. Or just "I called this, but I see that" situation (I bet 10$ this one will happen the next day)

### Authorisation

Even if it's just one partner backend-to-backend. It's **crucial** to have at least constant-in-the-code API key for them.

### Swagger

It's production if we can explain this protocol to those who should use it. **And** communicate changes efficiently.
We should do swagger and ideally code-gen a server from it.

I personally like [grpc gateway](https://github.com/chestnut42/terraforming-mars-manager/tree/main/pkg/api) very much.
It allows to generate swagger from proto + gives all the grpc ecosystem. But it costs performance and complexity.

At least [openapi server/client](https://github.com/oapi-codegen/oapi-codegen) should be generated.

### Request Limit

`k8s` scales its load by CPU/Memory. That holds a **MAJOR** problem when scaling up. If a sudden load spike arrives existing replicas
will die. `k8s` will then add more and more replicas. **BUT** in the end there will be tons of replicas that can't stand up.

Solution for the problem is to return 429 when the memory usage hits a certain limit. On my experience that worked very well.

If your system scales with the number of requests and can actually protect this service and swallow excessive calls - we can skip that.

### Test ALL bad input cases

No matter what tests we use for that. It's important to make sure it's actually impossible to submit bad medication.

- **(Preferred)** We can write [httptest](https://pkg.go.dev/net/http/httptest) tests [here](/internal/transport/http/medication/create_medication_test.go)
- We can just add more test cases for integration tests

### Tracing (Nice-to-have can be added after release)

On my experience it's just enough to add some `TraceID` to all logs. It's easy to do inside 
[logging middleware](/internal/utils/httpx/log.go)

This way you can quickly search all logs for the one given request.
