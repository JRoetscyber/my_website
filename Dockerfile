# Use a lightweight Python image for ARM architecture (Raspberry Pi)
FROM python:3.11-slim-bookworm

# Set the working directory in the container
WORKDIR /app

# Copy the requirements file into the container
COPY requirements.txt .

# Install system dependencies required for building some Python packages
RUN apt-get update && apt-get install -y --no-install-recommends build-essential && rm -rf /var/lib/apt/lists/*

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the application code into the container
COPY . .

# Ensure the database directory exists
RUN mkdir -p instance

# Expose the port the Flask app will run on
EXPOSE 6010

# Command to run the application using Gunicorn
# Using 4 workers, binding to 0.0.0.0:6010, and using the wsgi.py module
CMD ["gunicorn", "--workers=4", "--bind=0.0.0.0:6010", "wsgi:app"]