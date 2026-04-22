from flask import Blueprint, render_template

booking_bp = Blueprint('booking', __name__)

@booking_bp.route('/book')
def book():
    return render_template('book.html')