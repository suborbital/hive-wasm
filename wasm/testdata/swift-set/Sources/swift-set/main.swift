import Suborbital

class SwiftSet: Suborbital.Runnable {
    func run(input: String) -> String {
        Suborbital.CacheSet(key: "important", value: input, ttl: 0)

        return "hello"
    }
}

@_cdecl("init")
func `init`() {
    Suborbital.Set(runnable: SwiftSet())
}