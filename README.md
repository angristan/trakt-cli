# trakt-cli

```

████████╗██████╗  █████╗ ██╗  ██╗████████╗     ██████╗██╗     ██╗
╚══██╔══╝██╔══██╗██╔══██╗██║ ██╔╝╚══██╔══╝    ██╔════╝██║     ██║
   ██║   ██████╔╝███████║█████╔╝    ██║       ██║     ██║     ██║
   ██║   ██╔══██╗██╔══██║██╔═██╗    ██║       ██║     ██║     ██║
   ██║   ██║  ██║██║  ██║██║  ██╗   ██║       ╚██████╗███████╗██║
   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝        ╚═════╝╚══════╝╚═╝
```

This is a CLI for [trakt.tv](https://trakt.tv) using the [trakt.tv API](https://trakt.docs.apiary.io/).

![](https://user-images.githubusercontent.com/11699655/154494260-d3ff23ec-72b2-45e4-9f39-41f52119621b.png)

## Installation

Grab a binary build from the [releases](https://github.com/angristan/trakt-cli/releases).

## Development

```
git clone https://github.com/angristan/trakt-cli
cd trakt-cli
go build
```

## Usage

```
➜  trakt
Source code: https://github.com/angristan/trakt-cli

Usage:
  trakt-cli [command]

Available Commands:
  auth        Authenticate with trakt.tv
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  history     Show your watched history

Flags:
  -h, --help     help for trakt-cli

Use "trakt-cli [command] --help" for more information about a command.
```

## Authentication

You need to create a _Trakt API app_ to use the API.

Go to https://trakt.tv/oauth/applications/new and create a new app.

This will give you a _Client ID_ and _Client secret_ for your app.

You can now log in with the CLI:

```
➜  trakt auth --client-id xxx --client-secret yyy
Please go to https://trakt.tv/activate and enter the following code: XXXXXXXX
Successfully authenticated, creds written to ~/.trakt.yaml
```
