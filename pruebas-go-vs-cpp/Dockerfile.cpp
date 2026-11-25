FROM gcc:14.4.0 AS build
WORKDIR /src
COPY src/ /src/
RUN g++ -std=c++20 -O2 -static -pthread -o /servidor servidor.cpp && strip /servidor
FROM scratch
COPY --from=build /servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
