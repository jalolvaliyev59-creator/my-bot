import os
import logging
import asyncio
import uuid
from aiogram import Bot, Dispatcher, executor, types
from aiogram.types import ReplyKeyboardMarkup, KeyboardButton

# Tokenni faqat muhit o'zgaruvchisidan olamiz (Xavfsizlik uchun)
API_TOKEN = os.getenv("BOT_TOKEN")
if not API_TOKEN:
    raise ValueError("DIQQAT: BOT_TOKEN muhit o'zgaruvchisida topilmadi!")

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
    # Unikal ID orqali fayl nomlari to'qnashib ketishining oldini olamiz
    unique_name = str(uuid.uuid4())[:8]
    output_file = f"video_{message.from_user.id}_{unique_name}.mp4"
    
    try:
        process = await asyncio.create_subprocess_exec(
            "yt-dlp", "-f", "bestvideo+bestaudio/best", "--merge-output-format", "mp4", "-o", output_file, url,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        stdout, stderr = await process.communicate()
        
        if os.path.exists(output_file):
            with open(output_file, 'rb') as video_file:
                await message.reply_video(video_file, caption="🎬 Videongiz tayyor!")
        else:
            error_text = stderr.decode('utf-8', errors='ignore')[:300]
            await message.reply(f"❌ Yuklab bo'lmadi. Xatolik:\n<code>{error_text}</code>", parse_mode="HTML")
    except Exception as e:
        await message.reply(f"Xatolik yuz berdi: {e}")
    finally:
        if os.path.exists(output_file):
            os.remove(output_file)
        await bot.delete_message(message.chat.id, status_msg.message_id)

async def download_audio(message: types.Message, url: str):
    status_msg = await message.reply("⏳ Musiqa ajratib olinmoqda, iltimos kuting...")
    unique_name = str(uuid.uuid4())[:8]
    output_file = f"audio_{message.from_user.id}_{unique_name}.mp3"
    
    try:
        process = await asyncio.create_subprocess_exec(
            "yt-dlp", "-x", "--audio-format", "mp3", "-o", output_file, url,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        stdout, stderr = await process.communicate()
        
        if os.path.exists(output_file):
            with open(output_file, 'rb') as audio_file:
                await message.reply_audio(audio_file, caption="🎵 Musiqangiz tayyor!")
        else:
            error_text = stderr.decode('utf-8', errors='ignore')[:300]
            await message.reply(f"❌ Musiqani olib bo'lmadi. Xatolik:\n<code>{error_text}</code>", parse_mode="HTML")
    except Exception as e:
        await message.reply(f"Xatolik yuz berdi: {e}")
    finally:
        if os.path.exists(output_file):
            os.remove(output_file)
        await bot.delete_message(message.chat.id, status_msg.message_id)

if __name__ == '__main__':
    executor.start_polling(dp, skip_updates=True)
