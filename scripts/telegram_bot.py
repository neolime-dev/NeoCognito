#!/usr/bin/env python3
import os
import logging
import subprocess
from datetime import datetime
from telegram import Update
from telegram.ext import ApplicationBuilder, ContextTypes, CommandHandler, MessageHandler, filters
from dotenv import load_dotenv

# --- Configuration ---
# Load secrets from .env file
load_dotenv()

TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
ALLOWED_USER_ID = os.getenv("ALLOWED_USER_ID")
VAULT_PATH = os.path.expanduser("~/Vault")
INBOX_FILE = os.path.join(VAULT_PATH, "00_Inbox.md")
AUTOSAVE_SCRIPT = os.path.expanduser("~/.local/bin/autosave.sh")

# Logging Setup
logging.basicConfig(
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    level=logging.INFO
)

# --- Helper Functions ---

def is_authorized(user_id):
    """Check if the sender is the owner."""
    if not ALLOWED_USER_ID:
        logging.error("ALLOWED_USER_ID not set in .env")
        return False
    return str(user_id) == str(ALLOWED_USER_ID)

def run_autosave():
    """Triggers the autosave script to sync changes to Git."""
    try:
        subprocess.run([AUTOSAVE_SCRIPT], check=True, capture_output=True)
        logging.info("Autosave triggered successfully.")
    except Exception as e:
        logging.error(f"Failed to trigger autosave: {e}")

# --- Bot Handlers ---

async def start(update: Update, context: ContextTypes.DEFAULT_TYPE):
    user_id = update.effective_user.id
    if not is_authorized(user_id):
        logging.warning(f"Unauthorized access attempt from {user_id}")
        return

    await context.bot.send_message(
        chat_id=update.effective_chat.id,
        text="🧠 NeoCognito Link Online.\nSend me text to capture into Inbox."
    )

async def handle_text(update: Update, context: ContextTypes.DEFAULT_TYPE):
    user_id = update.effective_user.id
    if not is_authorized(user_id):
        return

    text = update.message.text
    timestamp = datetime.now().strftime("**%H:%M**")
    
    # Format: - **HH:MM** - [Mobile] Content
    entry = f"\n- {timestamp} - [Mobile] {text}"

    try:
        # 1. Append to Inbox
        with open(INBOX_FILE, "a", encoding="utf-8") as f:
            f.write(entry)
        
        logging.info(f"Captured: {text}")
        
        # 2. Feedback to User
        await update.message.reply_text("✅ Captured.")

        # 3. Trigger Sync
        run_autosave()

    except Exception as e:
        logging.error(f"Error writing to file: {e}")
        await update.message.reply_text(f"❌ Error: {e}")

# --- Main ---

if __name__ == '__main__':
    if not TOKEN:
        print("Error: TELEGRAM_BOT_TOKEN not found in .env")
        exit(1)
    
    application = ApplicationBuilder().token(TOKEN).build()
    
    start_handler = CommandHandler('start', start)
    text_handler = MessageHandler(filters.TEXT & (~filters.COMMAND), handle_text)
    
    application.add_handler(start_handler)
    application.add_handler(text_handler)
    
    print("NeoCognito Bot is running...")
    application.run_polling()
