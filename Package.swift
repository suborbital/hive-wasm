// swift-tools-version:5.3
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "suborbital",
    products: [
        .library(name: "suborbital", targets: ["suborbital"]),
    ],
    dependencies: [],
    targets: [
        .target(
            name: "suborbital",
            dependencies: [],
            path: "api/swift/Sources"),
    ]
)
