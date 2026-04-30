FROM python:3.10-slim

# 1. Install the C++ compiler (g++) so greenlet can build properly
RUN apt-get update && apt-get install -y g++

# 2. Upgrade pip to the latest version (sometimes this helps find better packages)
RUN pip install --upgrade pip

# 3. Create your user (if you used the previous example)
RUN useradd -m appuser
WORKDIR /app

# 4. Now install your Python packages
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
# Copy the rest of your application code
COPY . .

# Create the instance directory and give ownership to 'appuser'
RUN mkdir -p /app/instance && chown -R appuser:appuser /app/instance

# Switch to the non-root user for security
USER appuser

# Start Gunicorn directly
CMD ["gunicorn", "--workers=4", "--bind=0.0.0.0:6010", "wsgi:app"]