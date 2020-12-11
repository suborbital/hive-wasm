// swift-tools-version:5.3
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "Suborbital",
    products: [
        .library(name: "Suborbital", targets: ["SuborbitalLib"]),
    ],
    dependencies: [],
    targets: [
        .target(
            name: "SuborbitalLib",
            dependencies: [],
            path: "api/swift/Sources"),
    ]
)
