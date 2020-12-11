
class ExampleRunnable: Runnable {
    func run(input: String) -> String {
        log_info(msg: "testing")
        
        return "why hello " + input
    }
}

@_cdecl("init")
func `init`() {
    set(runnable: ExampleRunnable())
}