import os
import logging
import asyncio
import threading
from http.server import HTTPServer, BaseHTTPRequestHandler
from aiogram import Bot, Dispatcher, executor, types
from aiogram.types import ReplyKeyboardMarkup, KeyboardButton

API_TOKEN = os.getenv("BOT_TOKEN", "8510258357:AAFbZZOu1gmnGP34mPqLnxMpIS8-ibKtIpE")

logging.basicConfig(level=logging.INFO)

bot = Bot(token=API_TOKEN)
dp = Dispatcher(bot)

user_links = {}

keyboard = ReplyKeyboardMarkup(resize_keyboard=True)
keyboard.add(KeyboardButton("🎬 Videoni yuklash"), KeyboardButton("🎵 Musiqasini yuklash"))

@dp.message_handler(commands=['start', 'help'])
async def send_welcome(message: types.Message):
    await message.reply(
        "Assalomu alaykum! 🎬\n\nMenga Instagram, TikTok yoki YouTube havolasini yuboring:",
        reply_markup=keyboard
    )

@dp.message_handler(lambda message: message.text and (message.text.startswith("http://") or message.text.startswith("https://")))
async def receive_link(message: types.Message):
    user_links[message.from_user.id] = message.text
    await message.reply("Havola qabul qilindi! Nimani yuklab olishni xohlaysiz?", reply_markup=keyboard)

@dp.message_handler(lambda message: message.text in ["🎬 Videoni yuklash", "🎵 Musiqasini yuklash"])
async def handle_choice(message: types.Message):
    chat_id = message.from_user.id
    if chat_id not in user_links:
        await message.reply("❌ Avval video havolasini yuboring!")
        return
    
    url = user_links[chat_id]
    
    if message.text == "🎬 Videoni yuklash":
        await download_video(message, url)
    elif message.text == "🎵 Musiqasini yuklash":
        await download_audio(message, url)

async def download_video(message: types.Message, url: str):
    status_msg = await message.reply("⏳ Video yuklanmoqda, iltimos kuting...")
    output_file = f"video_{message.from_user.id}.mp4"
    
    try:
        process = await asyncio.create_subprocess_exec(
            "yt-dlp", "-f", "b[filesize<50M]/best", "-o", output_file, url,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        await process.communicate()
        
        if os.path.exists(output_file):
            await message.reply_video(open(output_file, 'rb'), caption="🎬 Videongiz tayyor!")
            os.remove(output_file)
        else:
            await message.reply("❌ Videoni yuklab bo'lmadi.")
    except Exception as e:
        await message.reply(f"Xatolik yuz berdi: {e}")
    finally:
        await bot.delete_message(message.chat.id, status_msg.message_id)

async def download_audio(message: types.Message, url: str):
    status_msg = await message.reply("⏳ Musiqa ajratib olinmoqda, iltimos kuting...")
    output_file = f"audio_{message.from_user.id}.mp3"
    
    try:
        process = await asyncio.create_subprocess_exec(
            "yt-dlp", "-x", "--audio-format", "mp3", "-o", output_file, url,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        await process.communicate()
        
        if os.path.exists(output_file):
            await message.reply_audio(open(output_file, 'rb'), caption="🎵 Musiqangiz tayyor!")
            os.remove(output_file)
        else:
            await message.reply("❌ Musiqani ajratib bo'lmadi.")
    except Exception as e:
        await message.reply(f"Xatolik yuz berdi: {e}")
    finally:
        await bot.delete_message(message.chat.id, status_msg.message_id)

# --- Render port talabini qondirish uchun oddiy HTTP server ---
class SimpleHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"Bot is running!")
    def log_message(self, format, *args):
        pass # Loglarni to'ldirmasligi uchun

def run_server():
    port = int(os.getenv("PORT", 8080))
    server = HTTPServer(("0.0.0.0", port), SimpleHandler)
    server.serve_forever()

if __name__ == '__main__':
    # Serverni alohida oqimda (thread) ishga tushiramiz
    server_thread = threading.Thread(target=run_server, daemon=True)
    server_thread.start()
    
    # Asosiy oqimda Telegram bot ishlaydi
    executor.start_polling(dp, skip_updates=True)
