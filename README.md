# gapu 

Small, fast, simple tool for retrieving all users and their information for public ArcGIS Portal groups.

You feed **GAPU** (Get ArcGIS Portal Users) the base rest URL, it returns all users that can be retrieved from an anonymous principal.

This can be a useful way of finding usernames belonging to a company using ArcGIS Portal without requiring any credentials.

## Installation

```sh
go get github.com/mirabis/gapu
```

## Usage
The most basic usage is to simply pass the rest portal URL to the -u parameter, for example:

```sh
mirabis~$ gapu -u https://maps.company.net/portal/sharing/rest/
ArcGIS group id, username, FullName, Join date timestamp
8f3e2a2430524f3ca47a03de7888e86e, renze.evert@contoso.com, "renze.evert", 1567750584798 
8f3e2a2430524f3ca47a03de7888e86e, rik.pauli@contoso.com, "Rik.pauli", 1554116686967 
8f3e2a2430524f3ca47a03de7888e86e, riva.ehyr@contoso.com, "Riva.ehyr", 1591080285396 
...
```

### Parameters

```sh
mirabis~$ gomu -h

Usage:
  gapu [OPTIONS]

 ██████╗  █████╗ ██████╗ ██╗   ██╗
██╔════╝ ██╔══██╗██╔══██╗██║   ██║
██║  ███╗███████║██████╔╝██║   ██║
██║   ██║██╔══██║██╔═══╝ ██║   ██║
╚██████╔╝██║  ██║██║     ╚██████╔╝
 ╚═════╝ ╚═╝  ╚═╝╚═╝      ╚═════╝ 
                                  
Application Options: (/* windows, -* Unix)
  /t, /threads:   Number of concurrent threads (default: 20)
  /u, /url:       ArcGIS Portal REST URL
                  (https://maps.company.net/portal/sharing/rest/)
  /a, /agent:
  /v, /verbose    Turns on verbose logging
  /e, /esri       Will also print default ESRI users
      /insecure   Switches all HTTPS calls to HTTP

Help Options:
  /?              Show this help message
  /h, /help       Show this help message
```


## Credits
- [hakluke](https://twitter.com/hakluke) my inspiration to start transitioning from Python/.NET to golang
- [tomnomnom](https://github.com/tomnomnom) my inspiration to start transitioning from Python/.NET to golang

### Contribution & License
You can contribute in following ways:

- Report bugs
- Give suggestions to make it better (I'm new to golang)
- Fix issues & submit a pull request

Do you want to have a conversation in private? Hit me up on my [twitter](https://twitter.com/iMirabis/), inbox is open :)

**gapu** is licensed under [GPL v3.0 license](https://www.gnu.org/licenses/gpl-3.0.en.html)