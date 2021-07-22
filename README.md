# Workshop: Building data-intensive apps with SingleStore, Redpanda, and Golang

This repo provides a starting point for building a data-intensive application
using SingleStore, Vectorized Redpanda, and Golang. SingleStore is a scale-out
relational database built for data-intensive workloads. Redpanda is a Kafka API
compatible streaming platform for mission-critical workloads created by the team
at Vectorized.

The rest of this readme will walk you through building a simple
data-intensive application starting from the code here. If you follow this workshop, you will
accomplish the following tasks:

1. Prepare your environment
2. Write a digital-twin
3. Define a schema and load the data using Pipelines
4. Expose business logic via an HTTP API
5. Visualize your data

## 1. Prepare your environment

Before we can start writing code, we need to make sure that your environment is
setup and ready to go.

1. Make sure you clone this git repository to your machine

   ```bash
   git clone https://github.com/singlestore-labs/singlestore-workshop-data-intensive-app.git
   ```

2. This project uses docker & docker-compose
   - [docker][docker]
   - [docker-compose][docker-compose]

> ❗ **Note:** If you are following this tutorial on Windows or Mac OSX you may
> need to increase the amount of RAM and CPU made available to docker. You can
> do this from the docker configuration on your machine. More information:
> [Mac OSX documentation][docker-ramcpu-osx],
> [Windows documentation][docker-ramcpu-win]

> ❗ **Note 2:** If you are following this tutorial on a Mac with the new M1
> chipset, it is unlikely to work. SingleStore does not yet support running on
> M1. We are actively fixing this, but in the meantime please follow along using
> a Linux virtual machine.

3. A code editor that you are comfortable using, if you aren't sure pick one of
   the options below:
   - [Visual Studio Code][vscode]

     > _Recommended:_ this repository is setup with VSCode launch configurations
     > as well as [SQL Tools][sqltools] support.

   - [Jetbrains Goland][jetbrains-go]

4. [Sign up][singlestore-signup] for a free SingleStore license. This allows you
   to run up to 4 nodes up to 32 gigs each for free. Grab your license key from
   [SingleStore portal][singlestore-portal] and set it as an environment
   variable.

    ```bash
    export SINGLESTORE_LICENSE="<<singlestore license>>"
    ```

### Test that your environment is working

> ℹ️ **Note:** This repository uses a bash script `./tasks` to run common tasks.
> If you run the script with no arguments it will print out some information
> about how to use it.

Before proceeding, please execute `./tasks up` in the root of this repository to
boot up all of the services we need and check that your environment is working
as expected.

```bash
$ ./tasks up
(... lots of docker output here ...)
   Name                  Command               State                                                                     Ports
-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
grafana       /run.sh                          Up      0.0.0.0:3000->3000/tcp,:::3000->3000/tcp
prometheus    /bin/prometheus --config.f ...   Up      0.0.0.0:9090->9090/tcp,:::9090->9090/tcp
redpanda      /bin/bash -c rpk config se ...   Up      0.0.0.0:29092->29092/tcp,:::29092->29092/tcp, 0.0.0.0:9092->9092/tcp,:::9092->9092/tcp, 0.0.0.0:9093->9093/tcp,:::9093->9093/tcp, 9644/tcp
simulator     ./simulator --config confi ...   Up
singlestore   /startup                         Up      0.0.0.0:3306->3306/tcp,:::3306->3306/tcp, 3307/tcp, 0.0.0.0:8080->8080/tcp,:::8080->8080/tcp
```

If the command fails or doesn't end with the above status report, you may need
to start debugging your environment. If you are completely stumped please file
an issue against this repository and someone will try to help you out.

If the command succeeds, then you are good to go! Lets check out the various
services before we continue:

| service            | url                   | port | username | password |
| ------------------ | --------------------- | ---- | -------- | -------- |
| SingleStore Studio | http://localhost:8080 | 8080 | root     | root     |
| Grafana Dashboards | http://localhost:3000 | 3000 | root     | root     |
| Prometheus         | http://localhost:9090 | 9090 |          |          |

You can open the three urls directly in your browser and login using the
provided username & password.

In addition, you can also connect directly to SingleStore DB using any MySQL
compatible client or [VSCode SQL Tools][sqltools]. For example:

```
$ mariadb -u root -h 127.0.0.1 -proot
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 28
Server version: 5.5.58 MemSQL source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> select "hello world";
+-------------+
| hello world |
+-------------+
| hello world |
+-------------+
1 row in set (0.001 sec)
```

To make sure that Redpanda is receiving data lets tail a test topic and
see what the hello simulator is saying:

```bash
$ ./tasks rpk topic consume --offset latest test
{
 "message": "{\"message\":\"hello world\",\"time\":\"2021-07-21T21:25:50.708768659Z\",\"worker\":\"sim-0\"}\n",
 "partition": 3,
 "offset": 2,
 "timestamp": "2021-07-21T21:25:50.708Z"
}
{
 "message": "{\"message\":\"hello world\",\"time\":\"2021-07-21T21:25:51.70910547Z\",\"worker\":\"sim-0\"}\n",
 "partition": 4,
 "offset": 2,
 "timestamp": "2021-07-21T21:25:51.709Z"
}
```

## 2. Write a digital-twin

Now that we have our environment setup it's time to start writing some code. Our
first task is to build a digital-twin, which is basically a data simulator. We
will use this digital-twin to generate large volumes of real-time data we need
to make our application truly _data-intensive_.

In this workshop, our digital-twin will be simulating simple page events for the
SingleStore website. Lets define it's behavior:

- The simulation will maintain a set of users browsing the website
- New users will be randomly created
- Users will "enter" the website at a random point in the sitemap
- Users will navigate through the tree via adjacent nodes (up, down, sibling) at
  random
- Users will spend a random amount of time on each page
- While users remain on the page we will periodically record how long they have
  been there
- No event will be generated when a user decides to leave the website

> _Note:_ This simulation is far from realistic and will result in extremely
> noisy data. A more sophisticated solution would be to take real page view data
> and use it to generate more accurate distributions (or build an AI replicating
> user behavior, left as an exercise to the reader). But for the purpose of this
> workshop, we won't worry too much about the quality of the data.

> _Note 2:_ I have already provided a **ton** of code and helpers to make this
> workshop run smoothly, so if it feels like something is a bit magic - I
> probably did that on purpose to keep the tutorial running smoothly. Please
> take a look through some of the provided helpers if you want to learn more!

### Define our data structures

First, we need to define an object to track our "users". I suggest doing this in
[simulator.go](src/simulator.go), but any go file in that directory (or a new go
file) will work fine.

```golang
type User struct {
  UserID      string
  CurrentPage *Page
  LastChange  time.Time
}
```

We will also need an event object which will help us write JSON to the events
topic in Redpanda.

```golang
type Event struct {
  Timestamp int64   `json:"unix_timestamp"`
  PageTime  float64 `json:"page_time_seconds,omitempty"`
  Referrer  string  `json:"referrer,omitempty"`
  UserID    string  `json:"user_id"`
  Path      string  `json:"path"`
}
```

As you can tell from the code, most of the boilerplate has already been written
for you. We mainly need to fill in the body of the `Simulator.Run` method. Lets
do that now.

### Data Structures

First we need a place to store our users. Lets do that at the start of the
`Run()` method. We will use a linked list data structure since we will be adding
and removing users often.

```golang
func (s *Simulator) Run() error {
  users := list.New()
```

We also need a way to write events to a Redpanda topic. For that we will create
a `TopicEncoder` which is a simple object which wraps a Kafka producer with a
JSON encoder.

```golang
events := s.producer.TopicEncoder("events")
```

### Creating users

Now it's time to start working in the main loop of the `Run()` method. Each time the
simulator ticks (once per second) we will create a random number of new users.

I have already provided a variable in `s.config` called `MaxUsersPerTick` which
we can use to play with later on. For now, write some code that looks a bit like
this:

```golang
for s.Running() {
    time.Sleep(time.Second)
    unixNow := time.Now().Unix()

    // create a random number of new users
    numNewUsers := int(math.Max(1, float64(s.config.MaxUsersPerTick)*rand.Float64()))
    for i := 0; i < numNewUsers; i++ {
        // define a new user
        user := &User{
            UserID:      NextUserId(),
            CurrentPage: s.sitemap.RandomLeaf(),
            LastChange:  time.Now(),
        }

        // add the user to the list
        users.PushBack(user)

        // write an event to the topic
        err := events.Encode(Event{
            Timestamp: unixNow,
            UserID:    user.UserID,
            Path:      user.CurrentPage.Path,

            // we provide a fake referrer here
            Referrer:  RandomReferrer(),
        })
        if err != nil {
            return err
        }
    }
}
```

### Simulating browsing activity

Now for the meat, we need to make our users browse the site! Lets add something
like this right after we create new users:

```golang
// we will be removing elements from the list while we iterate, so we
// need to keep track of next outside of the loop
var next *list.Element

// iterate through the users list and simulate each users behavior
for el := users.Front(); el != nil; el = next {
    // bookkeeping
    next = el.Next()
    user := el.Value.(*User)
    pageTime := time.Since(user.LastChange)

    // users only consider leaving a page after at least 5 seconds
    if pageTime > time.Second*5 {

        // eventProb is a random value from 0 to 1 but is weighted
        // to be closer to 0 most of the time
        eventProb := math.Pow(rand.Float64(), 2)

        if eventProb > 0.9 {
            // user has left the site
            users.Remove(el)
            continue
        } else if eventProb > 0.8 {
            // user jumps to a random page
            user.CurrentPage = s.sitemap.RandomLeaf()
            user.LastChange = time.Now()
        } else if eventProb > 0.5 {
            // user goes to the "next" page
            user.CurrentPage = user.CurrentPage.RandomNext()
            user.LastChange = time.Now()
        }
    }

    // write an event to the topic recording the time on the current page
    // note that if the user has changed pages above, that fact will be reflected here
    err := events.Encode(Event{
        Timestamp: unixNow,
        UserID:    user.UserID,
        Path:      user.CurrentPage.Path,
        PageTime:  pageTime.Seconds(),
    })
    if err != nil {
        return err
    }
}
```

Sweet! You have built your first digital twin! You can test it by running the
following commands:

```bash
$ ./tasks simulator
...output of building and running the simulator...

$ ./tasks logs simulator
Attaching to simulator
simulator      | 2021/07/22 03:52:16 starting simulation with 1 workers

$ ./tasks rpk topic consume --offset latest events
{
 "message": "{\"unix_timestamp\":1626925973,\"page_time_seconds\":8.305101859,\"user_id\":\"a1e684e0-2a8f-48f3-8af3-26e55aadb86b\",\"path\":\"/blog/case-study-true-digital-group-helps-to-flatten-the-curve-with-memsql\"}",
 "partition": 1,
 "offset": 3290169,
 "timestamp": "2021-07-22T03:52:53.9Z"
}
{
 "message": "{\"unix_timestamp\":1626925973,\"page_time_seconds\":1.023932243,\"user_id\":\"a2e684e0-2a8f-48f3-8af3-26e55aadb86b\",\"path\":\"/media-hub/releases/memsql67\"}",
 "partition": 1,
 "offset": 3290170,
 "timestamp": "2021-07-22T03:52:53.9Z"
}
...tons of messages, ctrl-C to cancel...
```

## 3. Define a schema and load the data using Pipelines

Next on our TODO list is getting the data into SingleStore! You can put away
your Go skills, this step is all about writing SQL.

> ℹ️ **Note:** Everyone has a different method of iteratively working on a
> schema, but for this workshop I recommend working inside the provided
> [schema.sql](schema.sql) file. The first line of the file is
> `drop database if exists app` which will **delete** the entire `app` database
> each time you reload this file. This makes iteration very fast, but just be
> careful not to run the file on a database full of juicy user data 😉.

So, lets start with a simple events table. Something like this should do the
trick:

```sql
CREATE TABLE events (
    ts DATETIME NOT NULL,
    path TEXT NOT NULL,
    user_id TEXT NOT NULL,

    referrer TEXT,
    page_time_s DOUBLE NOT NULL DEFAULT 0,

    SORT KEY (ts),
    SHARD KEY (user_id),

    KEY (user_id),
    KEY (referrer)
);
```

Next, we are going to use SingleStore Pipelines to consume the events topic we
created earlier. This is surprisingly easy, check it out:

```sql
CREATE PIPELINE events
AS LOAD DATA KAFKA 'redpanda/events'
SKIP DUPLICATE KEY ERRORS
INTO TABLE events
FORMAT JSON (
    @unix_timestamp <- unix_timestamp,
    path <- path,
    user_id <- user_id,
    referrer <- referrer DEFAULT NULL,
    page_time_s <- page_time_seconds DEFAULT 0
)
SET ts = FROM_UNIXTIME(@unix_timestamp);

START PIPELINE events;
```

The pipeline defined above connects to redpanda and starts loading the events
topic in batches. Each event is processed by the `FORMAT JSON` clause which
massages the event into the schema of the `events` table we defined above. You
can read more about the powerful `CREATE PIPELINE` command
[in our docs][create-pipeline].

## 4. Expose business logic via an HTTP API

Now that we have successfully generated and loaded data into SingleStore, we can
easily expose that data via a simple HTTP API. Open up [api.go](src/api.go) and define some interesting queries.

The first query I suggest writing is a simple leaderboard. It looks something
like this:

```sql
SELECT path, COUNT(DISTINCT user_id) AS count
FROM events
GROUP BY 1
ORDER BY 2 DESC
LIMIT 10
```

We can expose the results of this query via the following function added to
[api.go](src/api.go):

```golang
func (a *Api) Leaderboard(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an int"})
		return
	}

	out := []struct {
		Path  string `json:"path"`
		Count int    `json:"count"`
	}{}

	err := a.db.SelectContext(c.Request.Context(), &out, `
		SELECT path, COUNT(DISTINCT user_id) AS count
		FROM events
		GROUP BY 1
		ORDER BY 2 DESC
		LIMIT ?
	`, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}
```

Then register the function in the `RegisterRoutes` function like so:

```golang
func (a *Api) RegisterRoutes(r *gin.Engine) {
	r.GET("/ping", a.Ping)
	r.GET("/leaderboard", a.Leaderboard)
}
```

You can run the api and test your new endpoint using the following commands:

```bash
$ ./tasks api
...output of building and running the api...

$ ./tasks logs api
Attaching to api
api            | [GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.
api            |
api            | [GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
api            |  - using env:  export GIN_MODE=release
api            |  - using code: gin.SetMode(gin.ReleaseMode)
api            |
api            | [GIN-debug] GET    /ping                     --> src.(*Api).Ping-fm (3 handlers)
api            | [GIN-debug] GET    /leaderboard              --> src.(*Api).Leaderboard-fm (3 handlers)
api            | [GIN-debug] Listening and serving HTTP on :8000

$ curl -s "localhost:8000/leaderboard?limit=2" | jq
[
  {
    "path": "/blog/memsql-spark-connector",
    "count": 524
  },
  {
    "path": "/blog",
    "count": 507
  }
]
```

Before moving forward consider creating one or two additional endpoints or
modifying the leaderboard. Here are some ideas:

- (_easy_): change the leaderboard to accept a filter for a specific path
  prefix; i.e. `-/leaderboard?prefix=/blog`
- (_medium_): add a new endpoint which returns a referrer leaderboard (you only
  care about rows where referrer is `NOT NULL`)
- (_hard_): add a new endpoint which returns the number of page loads over time
  bucketed by minute - keep in mind that each row in the events table represents
  a user viewing a page for 1 second

## 5. Visualize your data

Whew! We are almost at the end of the workshop! As a final piece of the puzzle,
lets visualize our data using Grafana. I have already setup grafana for you, so
just head over to http://localhost:3000 to get started. Username and password
are both `root`.

Assuming you have ended up with a mostly working simulation, you can check out
the dashboard I have already created here:
http://localhost:3000/d/_TsB4vZ7k/user-analytics?orgId=1&refresh=5s

I highly recommend playing around with Grafana and experimenting with it's many
features. I have setup SingleStore as a datasource called "SingleStore" so it
should be pretty easy to create new dashboards and panels.

For further Grafana learning, I recommend checking out their docs
[starting here][grafana-getting-started]. Good luck!

<!-- Link index -->

[docker-compose]: https://docs.docker.com/compose/install/
[docker-ramcpu-osx]: https://docs.docker.com/docker-for-mac/#resources
[docker-ramcpu-win]: https://docs.docker.com/docker-for-windows/#resources
[docker]: https://docs.docker.com/get-docker/
[jetbrains-go]: https://www.jetbrains.com/go/
[sqltools]: https://marketplace.visualstudio.com/items?itemName=mtxr.sqltools
[vscode]: https://code.visualstudio.com/
[singlestore-signup]: https://www.singlestore.com/try-free/
[singlestore-portal]: https://portal.singlestore.com/?utm_medium=osm&utm_source=github
[create-pipeline]: https://docs.singlestore.com/db/v7.3/en/reference/sql-reference/pipelines-commands/create-pipeline.html
[grafana-getting-started]: https://grafana.com/docs/grafana/latest/getting-started/getting-started/
