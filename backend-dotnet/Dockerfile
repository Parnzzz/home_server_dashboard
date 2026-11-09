FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

COPY BackendDotnet.csproj ./
RUN dotnet restore BackendDotnet.csproj

COPY . ./
RUN dotnet publish BackendDotnet.csproj --configuration Release --output /app/publish --no-restore

FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS runtime
WORKDIR /app
ENV ASPNETCORE_URLS=http://0.0.0.0:8081
EXPOSE 8081

COPY --from=build /app/publish ./
ENTRYPOINT ["dotnet", "BackendDotnet.dll"]
