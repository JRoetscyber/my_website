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

# Create instance directory and set open permissions so SQLite can write to it
# even after Docker mounts the named volume over it
RUN mkdir -p /app/instance && chmod 777 /app/instance

# Expose the port the Flask app will run on
EXPOSE 6010

# Use an entrypoint script so the instance dir permissions are fixed at runtime
# (Docker volumes mount AFTER the image layers, so we fix perms on startup)
CMD ["sh", "-c", "mkdir -p /app/instance && chmod 777 /app/instance && gunicorn --workers=4 --bind=0.0.0.0:6010 wsgi:app"]