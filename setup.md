# Setup Guide

Follow these steps to set up the Bill Processing System locally:

## 1. Clone the Repository

Clone the repository to your local machine:

```sh
git clone <repo-url>
cd Bill-Processing-System
```

## 2. Install Temporal

- Visit [Temporal Downloads](https://docs.temporal.io/docs/server/quick-install/) and follow the instructions for your OS.
- For Windows, you can use Docker Desktop and run:

```sh
docker run --rm -d -p 7233:7233 --name temporal \
  temporalio/auto-setup:latest
```

Or, if you have the Temporal CLI installed:

```sh
temporal server start-dev
```

## 3. Install Encore

- Visit [Encore.dev Install](https://encore.dev/docs/install) and follow the instructions for your OS.
- For Windows (PowerShell):

```powershell
iwr -useb https://encore.dev/install.ps1 | iex
```

## 4. Start Temporal Server

In one terminal, start the Temporal development server:

```sh
temporal server start-dev
```

## 5. Start Encore

In another terminal, run the Encore app:

```sh
encore run
```

## 6. Run Encore Unit Tests

To run Encore-based unit tests (e.g., for the `bills` package):

```sh
encore test ./bills/
```
