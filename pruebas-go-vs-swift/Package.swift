// swift-tools-version:6.0
import PackageDescription

let package = Package(
    name: "Servidor",
    platforms: [.macOS(.v13)],
    dependencies: [
        .package(url: "https://github.com/vapor/vapor.git", exact: "4.122.3"),
    ],
    targets: [
        .executableTarget(name: "Servidor", dependencies: [.product(name: "Vapor", package: "vapor")]),
    ]
)
