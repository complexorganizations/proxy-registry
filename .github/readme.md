# Proxy-registry

[![Updating the resources](https://github.com/complexorganizations/proxy-registry/actions/workflows/auto-update-repo.yml/badge.svg)](https://github.com/complexorganizations/proxy-registry/actions/workflows/auto-update-repo.yml)

A public registry of working proxy.

***Proxy-Directory-Registry is not yet complete. You should not rely on this code. It has not undergone proper degrees of security auditing and the protocol is still subject to change. We're working toward a stable 1.0.0 release, but that time has not yet come. There are experimental snapshots tagged with "0.0.0.MM-DD-YYYY", but these should not be considered real releases and they may contain security vulnerabilities (which would not be eligible for CVEs, since this is pre-release snapshot software). If you are packaging Proxy-Directory-Registry, you must keep up to date with the snapshots.***


## Features

- Scraping proxy lists
- Validation proxy lists
- Testing proxy lists


## Updating

Clone the latest version of the repo using git.
```
git clone https://github.com/complexorganizations/proxy-registry
```
Go to the directory.
```
cd proxy-registry/
```
Build the application.
```
go build .
```
Run the application
```
./proxy-registry -update
```


## FAQ

#### How often is this updated?

Everyday at 00:00 UTC.


## Contributing

Contributions are always welcome!

See `.github/contributing.md` for ways to get started.

Please adhere to this project's `.github/code_of_conduct.md`.


## Roadmap

- 

- 


## Authors

- [@prajwal-koirala](https://github.com/prajwal-koirala)


## Support

Please utilize the github repo issue and wiki for help.


## Feedback

Please utilize the github repo conversations to offer feedback.


## License

[Apache License Version 2.0](https://github.com/complexorganizations/proxy-directory-registry/blob/main/.github/license)
