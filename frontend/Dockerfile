FROM oven/bun:1

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

COPY . .
RUN chmod +x node_modules/.bin/*

EXPOSE 5173

CMD ["bun", "run", "dev", "--host"]
