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
   A story generator that starts with a random title based on user input. The user then makes decisions throughout the story, influencing the final outcome like a "butterfly effect.". This version includes a full audio rendition of the story once completed (using TTS Model), as well as a custom cover image generated by an LLM image model (DALL-E).

4. **Creative Content Generator**  
   This tool allows users to upload an image, which is analyzed to inspire various creative content options, such as poems, monologues, or short stories. The generated text can also be rendered into an audio format using LLM TTS (text-to-speech), and a custom cover image for the content is generated with an LLM image generator (DALL-E).

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

Berikut adalah versi README yang telah diperbarui sesuai dengan informasi tambahan yang Anda berikan:

---

# Web-Based Scraping Project and LLM Integration using Claude & GPT

This repository showcases a collection of mini-projects that integrate web scraping with Large Language Models (LLMs) such as Claude and GPT. These projects demonstrate how data from various sources can be collected, analyzed, and transformed using AI models for both functional and creative applications.

## Overview

This project includes several mini-applications, each with unique use cases and capabilities. Some applications now feature enhanced AI functionality, such as text-to-speech and image generation.

### Mini Projects

1. **Medium Account Roast**  
   A tool that scrapes content from a Medium account and uses LLMs to generate a humorous "roast" or analytical commentary on the material.

2. **Website Roast/Analysis**  
   This project scrapes a target website to retrieve content based on user-selected topics. Claude or GPT is then used to conduct a humorous roast or detailed analysis, providing insights or an entertaining review of the chosen topic.

3. **Butterfly Effect Stories Generator**  
   A dynamic story generator where the user begins with a random title based on their input, making choices throughout to influence the outcome. This version now includes a full audio rendition of the story once completed, as well as a custom cover image generated by an LLM image model (e.g., DALL-E).

4. **Creative Content Generator**  
   This new tool allows users to upload an image, which is analyzed to inspire various creative content options, such as poems, monologues, or short stories. The generated text can also be rendered into an audio format using LLM TTS (text-to-speech), and a custom cover image for the content is generated with an LLM image generator (e.g., DALL-E).

---

Each of these projects combines the power of web scraping with LLM-driven creativity and analysis, offering a wide array of applications from data-based analysis to generating personalized and entertaining content.