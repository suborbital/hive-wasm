import Suborbital

class FetchSwift: Suborbital.Runnable {
    func run(input: String) -> String {
        return Suborbital.HttpGet(url: input)
    }
}

@_cdecl("init")
func `init`() {
    Suborbital.Set(runnable: FetchSwift())
}