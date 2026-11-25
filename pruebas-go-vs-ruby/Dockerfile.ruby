FROM ruby:4.0.9-slim
RUN apt-get update && apt-get install -y --no-install-recommends build-essential && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY app/Gemfile /app/
RUN bundle install --jobs 4 && apt-get purge -y build-essential && apt-get autoremove -y
COPY app/config.ru app/puma.rb /app/
EXPOSE 8080
ENV WORKERS=0
ENTRYPOINT ["bundle", "exec", "puma", "-C", "puma.rb", "config.ru"]
