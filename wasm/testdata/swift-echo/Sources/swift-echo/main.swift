import Suborbital

class SwiftEcho: Suborbital.Runnable {
    func run(input: String) -> String {
        Suborbital.LogInfo(msg: "input: \(input.utf8)")

        let method = Suborbital.ReqMethod()
        let url = Suborbital.ReqURL()
        let helloTo = Suborbital.State(key: "hello")
        let baz = Suborbital.ReqBodyField(key: "baz") //testing it doesn't crash when something doesn't exist
        
        Suborbital.LogInfo(msg: "url: \(url)")
        Suborbital.LogInfo(msg: "method: \(method)")
        Suborbital.LogInfo(msg: "helloTo: \(helloTo)")
        Suborbital.LogInfo(msg: "baz: \(baz)")

        return "hello " + helloTo
    }
}

@_cdecl("init")
func `init`() {
    Suborbital.Set(runnable: SwiftEcho())
}