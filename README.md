# Web Based Scrapping Project and LLM Integration using Claude & GPT

This repository contains a collection of mini-projects that showcase implementation of web scraping  combined with the integration of Large Language Models (LLMs) like Claude and GPT. The purpose of this project is to demonstrate how data from various sources can be scraped and further analyzed or transformed using AI models.

## Overview

The project includes several mini-applications:

### Mini Projects
1. **Medium Account Roast**  
   A tool that scrapes data from a Medium account and generates a humorous "roast" or analysis of the content using LLMs.
   
2. **Website Roast/Analysis**  
   This mini-project scrapes a website to get some topic inside it and with topic user select, do a roast or detailed analysis using Claude or GPT, offering insights or Entertaining Review of the topic.

3. **Butterfly Effect Stories Generator**  
   A story generator that starts with a random title based on user input. The user then makes decisions throughout the story, influencing the final outcome like a "butterfly effect."

## Getting Started

### 1. Configure Environment Variables
Create a `.env` file in the root directory of the project and fill it with your configuration settings with basic values from `.example.env`.

### 2. Install Dependencies
Run the following command to ensure all necessary modules are installed:

```bash
go mod tidy
```

### 3. Start the Development Server
To start the development server, run:

```bash
go run main.go
```

This will start the server and automatically load changes when you rerun the command after making changes.

Alternatively, if you prefer using air for live reloading during development, simply run:

```bash
air
```

Make sure to configure air according to your project's needs by adjusting the settings in the .air.toml file.

### 4. Start the Production Server
To start the server in production mode, you can build the binary and run it:

#### On Windows:
```bash
go build -o lorem.exe
lorem.exe
```

#### On Linux/macOS:
```bash
go build -o lorem
./lorem
```
