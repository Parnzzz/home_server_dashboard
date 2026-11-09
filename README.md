Home Server Dashboard

A self-hosted server monitoring platform designed to monitor and manage home servers. The system collects and displays real-time system metrics, including CPU, memory, disk usage, and server health.

Features

Server Monitoring — Monitor connected servers and their status.

CPU Monitoring — Track CPU usage.

Memory Monitoring — Track RAM usage and utilization.

Disk Monitoring — Monitor disk usage and available storage.

Real-time Metrics — Collect and visualize system performance metrics.

Server Management — Manage registered servers.

Authentication — Secure access to the dashboard.

Tech Stack

Technology

Purpose

Flutter

Frontend application

.NET / Go

Backend services

PostgreSQL

Database

Docker

Containerization

GitHub Actions

CI/CD

Project Structure

home-server-dashboard/
├── frontend/              # Flutter application
├── backend/               # Backend services
├── docker-compose.yml     # Development environment
├── docker-compose.demo.yml
├── docker-compose.prod.yml
├── .env.example           # Environment configuration template
└── README.md

Getting Started

Prerequisites

Flutter SDK

.NET SDK

Go

Docker & Docker Compose

PostgreSQL

Installation

Clone the repository.

git clone https://github.com/USERNAME/home-server-dashboard.git
cd home-server-dashboard

Configure environment variables.

cp .env.example .env

Update the .env file with your local configuration.

Start the services.

docker compose up -d --build

Run the Flutter application.

cd frontend
flutter pub get
flutter run

Environment Configuration

The project supports multiple environments:

Development — Local development and debugging.

Demo — Testing and demonstration.

Production — Live deployment.

See .env.example for the required environment variables.

Security

Never commit .env files or credentials.

Keep secrets and API keys outside the source code.

Use separate credentials for each environment.

Project Status

In Development

This project is actively being developed. Features and architecture may change over time.

License

To be determined.
