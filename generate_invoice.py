from fpdf import FPDF
from datetime import datetime, timedelta
import os

class PDF(FPDF):
    def __init__(self, company_info, branding, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.company_name = company_info['name']
        self.company_address = company_info['address']
        self.company_phone = company_info['phone']
        self.company_email = company_info['email']
        self.logo_path = branding['logo']
        self.brand_color = branding['colors']['core']
        self.text_color = branding['colors']['text']
        self.bg_color = branding['colors']['background']
        self.heading_font = branding['fonts']['heading'] if branding['fonts']['heading'] else 'Helvetica'
        self.body_font = branding['fonts']['body'] if branding['fonts']['body'] else 'Helvetica'

        self.set_auto_page_break(auto=True, margin=15)
        self.add_page()
        self.set_fill_color(*self._hex_to_rgb(self.bg_color))
        self.rect(0, 0, self.w, self.h, 'F')
        self.set_text_color(*self._hex_to_rgb(self.text_color))

    def _hex_to_rgb(self, hex_color):
        hex_color = hex_color.lstrip('#')
        return tuple(int(hex_color[i:i+2], 16) for i in (0, 2, 4))

    def header(self):
        # Set branding color for header text
        self.set_text_color(*self._hex_to_rgb(self.brand_color))

        # Logo
        if self.logo_path and os.path.exists(self.logo_path):
            self.image(self.logo_path, 10, 8, 33)
        else:
            self.set_font(self.heading_font, 'B', 16)
            self.cell(0, 10, self.company_name, 0, 1, 'L')
            self.ln(5)

        # Company Info (top right)
        self.set_xy(self.w - 80, 10)
        self.set_font(self.body_font, 'B', 10)
        self.cell(0, 5, self.company_name, 0, 1, 'R')
        self.set_x(self.w - 80)
        self.set_font(self.body_font, '', 9)
        self.cell(0, 5, self.company_address, 0, 1, 'R')
        self.set_x(self.w - 80)
        self.cell(0, 5, f"Phone: {self.company_phone}", 0, 1, 'R')
        self.set_x(self.w - 80)
        self.cell(0, 5, f"Email: {self.company_email}", 0, 1, 'R')
        self.ln(20) # Space after header

        # Reset text color for body
        self.set_text_color(*self._hex_to_rgb(self.text_color))

    def footer(self):
        self.set_y(-15)
        self.set_font(self.body_font, 'I', 8)
        self.set_text_color(*self._hex_to_rgb(self.brand_color)) # Footer text in brand color
        self.cell(0, 10, f'Page {self.page_no()}/{{nb}}', 0, 0, 'C')

    def add_section_title(self, title):
        self.set_font(self.heading_font, 'B', 12)
        self.set_text_color(*self._hex_to_rgb(self.brand_color))
        self.cell(0, 10, title, 0, 1, 'L')
        self.set_text_color(*self._hex_to_rgb(self.text_color))
        self.ln(2)

    def add_client_info(self, client_name, client_address, client_email):
        self.add_section_title("Bill To")
        self.set_font(self.body_font, 'B', 10)
        self.cell(0, 6, client_name, 0, 1, 'L')
        self.set_font(self.body_font, '', 10)
        self.cell(0, 6, client_address, 0, 1, 'L')
        self.cell(0, 6, client_email, 0, 1, 'L')
        self.ln(10)

    def add_items_table(self, items, total_amount, total_time_spent):
        self.add_section_title("Invoice Items")
        self.set_font(self.body_font, 'B', 10)
        self.set_fill_color(*self._hex_to_rgb(self.brand_color))
        self.set_text_color(*self._hex_to_rgb(self.bg_color)) # White text on brand color for header

        # Table Header
        self.cell(80, 10, "Description", 1, 0, 'C', 1)
        self.cell(20, 10, "Qty", 1, 0, 'C', 1)
        self.cell(30, 10, "Unit Price", 1, 0, 'C', 1)
        self.cell(30, 10, "Time Spent (Hrs)", 1, 0, 'C', 1) # New column for time spent
        self.cell(30, 10, "Total", 1, 1, 'C', 1)

        self.set_font(self.body_font, '', 10)
        self.set_text_color(*self._hex_to_rgb(self.text_color)) # Reset text color for rows
        self.set_fill_color(255, 255, 255) # White fill for rows, adjust if different bg needed

        for item in items:
            # Multi_cell and then cell for remaining columns in the same row
            x_start = self.get_x()
            y_start = self.get_y()
            # Calculate height for multi_cell
            description_lines = self.multi_cell(80, 10, item['description'], 0, 'L', 0, 0, '', '', True)
            line_height = self.font_size * 1.2
            cell_height = max(10, description_lines * line_height) # Ensure min height of 10

            self.set_xy(x_start, y_start) # Reset position to draw border for multi_cell
            self.multi_cell(80, cell_height, item['description'], 1, 'L', 0, 0, '', '', True) # Draw multi_cell with border

            self.set_xy(x_start + 80, y_start) # Move to next column
            self.cell(20, cell_height, str(item['quantity']), 1, 0, 'C')
            self.cell(30, cell_height, f"${item['unit_price']:.2f}", 1, 0, 'R')
            self.cell(30, cell_height, str(item['time_spent']), 1, 0, 'C')
            self.cell(30, cell_height, f"${item['total']:.2f}", 1, 1, 'R') # 1,1 for newline
        
        # Totals
        self.set_fill_color(*self._hex_to_rgb(self.brand_color)) # Re-apply brand color fill
        self.set_text_color(*self._hex_to_rgb(self.bg_color)) # White text on brand color for totals
        self.cell(160, 10, "Total Time Spent (Hrs):", 1, 0, 'R', 1) # Highlight time total
        self.cell(30, 10, str(total_time_spent), 1, 1, 'C', 1)

        self.cell(160, 10, "TOTAL DUE:", 1, 0, 'R', 1)
        self.cell(30, 10, f"${total_amount:.2f}", 1, 1, 'R', 1)
        self.ln(10)

    def add_payment_details(self, payment_terms, bank_details):
        self.add_section_title("Payment Information")
        self.set_font(self.body_font, '', 10)
        self.cell(0, 6, f"Payment Terms: {payment_terms}", 0, 1, 'L')
        self.multi_cell(0, 6, f"Bank Details: {bank_details}", 0, 'L')
        self.ln(10)

    def add_notes(self, notes):
        self.add_section_title("Notes")
        self.set_font(self.body_font, '', 9)
        self.multi_cell(0, 5, notes)
        self.ln(10)

def generate_invoice(output_filename, invoice_details, company_info, branding, client_info, items):
    pdf = PDF(company_info, branding)

    # Add fonts - fpdf2 needs font files to embed custom fonts
    # For simplicity, if custom fonts are not found, it will fallback to Helvetica/Arial.
    # To use custom fonts, you would need .ttf files and add them like:
    # pdf.add_font('Syne', '', 'Syne-Regular.ttf', uni=True)
    # pdf.add_font('Inter', '', 'Inter-Regular.ttf', uni=True)
    # pdf.add_font('JetBrains Mono', '', 'JetBrainsMono-Regular.ttf', uni=True)
    # For now, we'll rely on basic font family names that FPDF knows or its fallbacks.

    pdf.add_page()
    pdf.set_text_color(*pdf._hex_to_rgb(branding['colors']['text']))

    # Invoice Number and Date
    pdf.set_font(pdf.heading_font, 'B', 20)
    pdf.set_text_color(*pdf._hex_to_rgb(pdf.brand_color))
    pdf.cell(0, 10, f"INVOICE #{invoice_details['invoice_number']}", 0, 1, 'L')
    pdf.set_text_color(*pdf._hex_to_rgb(pdf.text_color))
    pdf.set_font(pdf.body_font, '', 10)
    pdf.cell(0, 6, f"Date: {invoice_details['issue_date']}", 0, 1, 'L')
    pdf.cell(0, 6, f"Due Date: {invoice_details['due_date']}", 0, 1, 'L')
    pdf.ln(15)

    pdf.add_client_info(client_info['name'], client_info['address'], client_info['email'])

    total_amount = sum(item['total'] for item in items)
    total_time_spent = sum(item['time_spent'] for item in items)
    pdf.add_items_table(items, total_amount, total_time_spent)

    pdf.add_payment_details(invoice_details['payment_terms'], invoice_details['bank_details'])

    # Custom notes for invoice
    custom_notes = f"""Project Time Taken: {total_time_spent} hours.

{invoice_details['standard_notes']}"""
    pdf.add_notes(custom_notes)


    pdf.output(output_filename)

if __name__ == "__main__":
    # --- Configuration ---
    COMPANY_INFO = {
        'name': "JO4 Dev",
        'address': "20 west riding road, ferryval, nigel",
        'phone': "0826102440",
        'email': "jroetscyber@gmail.com"
    }

    BRANDING = {
        'logo': "C:\Usersoets\OneDrive\Apps\portfolio\my_website\apple-touch-icon.png", # Corrected logo path
        'colors': {
            'core': "#00e5cc",
            'background': "#000000",
            'text': "#f0f0f0"
        },
        'fonts': {
            'heading': "Syne", # FPDF uses generic names, actual font needs to be embedded
            'body': "Inter",   # See comments in generate_invoice function
            'monospace': "JetBrains Mono"
        }
    }

    INVOICE_DETAILS_TEMPLATE = {
        'invoice_number_prefix': datetime.now().strftime("INV-%y%m-"), # Prefix, will append a counter
        'issue_date': datetime.now().strftime("%Y-%m-%d"),
        'payment_terms_days': 14, # Store as int for calculation
        'payment_terms_text': "within 14 days",
        'bank_details': "First Business Zero Account, Acc: 63060057809, Branch: 250655 (FNB)",
        'standard_notes': "Thank you for your business! We appreciate your prompt payment."
    }

    # --- Interactive Input ---
    print("
--- Enter Client Information ---")
    client_name = input("Client Name: ")
    client_address = input("Client Address: ")
    client_email = input("Client Email: ")

    CLIENT_INFO = {
        'name': client_name,
        'address': client_address,
        'email': client_email
    }

    ITEMS = []
    print("
--- Enter Invoice Items (type 'done' for description to finish) ---")
    item_count = 1
    while True:
        print(f"
Item #{item_count}:")
        description = input("  Description (type 'done' to finish): ")
        if description.lower() == 'done':
            break

        while True:
            try:
                quantity = int(input("  Quantity: "))
                break
            except ValueError:
                print("Invalid quantity. Please enter a number.")

        while True:
            try:
                unit_price = float(input("  Unit Price: "))
                break
            except ValueError:
                print("Invalid unit price. Please enter a number.")

        while True:
            try:
                time_spent = float(input("  Time Spent (Hrs): "))
                break
            except ValueError:
                print("Invalid time spent. Please enter a number.")

        total = quantity * unit_price
        ITEMS.append({
            'description': description,
            'quantity': quantity,
            'unit_price': unit_price,
            'time_spent': time_spent,
            'total': total
        })
        item_count += 1

    if not ITEMS:
        print("No items entered. Invoice generation cancelled.")
    else:
        # Generate dynamic invoice number
        invoice_counter = 1
        # Construct base filename without counter for checking
        base_output_filename_prefix = f"JO4_Dev_Invoice_{INVOICE_DETAILS_TEMPLATE['invoice_number_prefix']}"
        while os.path.exists(f"{base_output_filename_prefix}{invoice_counter:03d}.pdf"):
            invoice_counter += 1
        final_invoice_number = f"{INVOICE_DETAILS_TEMPLATE['invoice_number_prefix']}{invoice_counter:03d}"

        # Calculate due_date
        due_date = datetime.now() + timedelta(days=INVOICE_DETAILS_TEMPLATE['payment_terms_days'])

        INVOICE_DETAILS = {
            'invoice_number': final_invoice_number,
            'issue_date': INVOICE_DETAILS_TEMPLATE['issue_date'],
            'due_date': due_date.strftime("%Y-%m-%d"),
            'payment_terms': INVOICE_DETAILS_TEMPLATE['payment_terms_text'],
            'bank_details': INVOICE_DETAILS_TEMPLATE['bank_details'],
            'standard_notes': INVOICE_DETAILS_TEMPLATE['standard_notes']
        }

        output_pdf_name = f"JO4_Dev_Invoice_{final_invoice_number}.pdf"
        generate_invoice(output_pdf_name, INVOICE_DETAILS, COMPANY_INFO, BRANDING, CLIENT_INFO, ITEMS)
        print(f"Invoice generated: {output_pdf_name}")