FROM python:3.9-slim

WORKDIR /app

# yt-dlp uchun ffmpeg va curl ni o'rnatish
RUN apt-get update && apt-get install -y ffmpeg curl && rm -rf /var/lib/apt/lists/*

# Pip'ni yangilash va yt-dlp ni to'g'ridan-to'g'ri o'rnatish
COPY requirements.txt .
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

# Qolgan barcha fayllarni ko'chirish
COPY . .

CMD ["python", "bot.py"]
