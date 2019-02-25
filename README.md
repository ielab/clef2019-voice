# CLEF2019 Query Voice Interface

## How to run

The server is written in Go. Therefore, to run the server, use:

```bash
$ go build
$ ./clef2019-voice
```

or just

```bash
$ go run *.go
```

The static and template files must be in the same directory as the binary is run (i.e., `web/`). 

## Adding topics

A folder called `topics` is scanned when the server is started to find topics. Topics are `.wav` audio files containing a spoken narrative. Each file name should be the id of the respective topic, followed by the `.wav` file extension (e.g., `1.wav`, `2.wav`, etc.).

## Managing participants

Access the admin console by navigating to `/admin/user`. The `/admin` route is protected by basic authentication. The username is always `admin` and the password can be configured in a file called `config.toml`. For example:

```toml
admin_password = "changeme"
```

From here, users can be added, removed, and assigned topics.

## Exporting data

Access the admin console by navigating to `/admin/data`. Previous exports are made visible as well as a button for exporting data in csv format. A breakdown of the overall progress and the progress of each participant is displayed below.

## Participating

Access the interface by navigating to `/voice`. Enter the assigned participant id in the box. From here, you are presented with a list of assigned topics. Click any of the links to complete a topic.

  