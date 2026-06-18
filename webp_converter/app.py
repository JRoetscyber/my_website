import subprocess
import os
from flask import Flask, request, render_template, jsonify

app = Flask(__name__)

UPLOAD_FOLDER = "uploads"
OUTPUT_FOLDER = "outputs"
ALLOWED = {"jpg", "jpeg", "png"}

def allowed(filename):
    return "." in filename and filename.rsplit(".", 1)[1].lower() in ALLOWED

@app.route("/")
def index():
    return render_template("index.html")

@app.route("/convert", methods=["POST"])
def convert():
    if "image" not in request.files:
        return jsonify({"error": "No file uploaded"}), 400

    file = request.files["image"]
    quality = request.form.get("quality", "80")
    method  = request.form.get("method",  "6")

    if not allowed(file.filename):
        return jsonify({"error": "Only JPG and PNG allowed"}), 400

    # Save uploaded file
    inputPath  = os.path.join(UPLOAD_FOLDER, file.filename)
    outputName = file.filename.rsplit(".", 1)[0] + ".webp"
    outputPath = os.path.join(OUTPUT_FOLDER, outputName)
    file.save(inputPath)

    # Call your C++ converter
    result = subprocess.run(
        ["./converter", inputPath, outputPath, quality, method],
        capture_output=True, text=True
    )

    if result.returncode != 0:
        return jsonify({"error": result.stderr}), 500

    # Parse stats from C++ output
    inputSize  = os.path.getsize(inputPath)
    outputSize = os.path.getsize(outputPath)
    saving     = round(100 * (1 - outputSize / inputSize), 2)

    return jsonify({
        "original_name"  : file.filename,
        "webp_name"      : outputName,
        "input_size_kb"  : round(inputSize  / 1024, 1),
        "output_size_kb" : round(outputSize / 1024, 1),
        "saving_percent" : saving,
        "quality"        : quality,
        "method"         : method,
        "converter_log"  : result.stdout
    })

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000, debug=True)
