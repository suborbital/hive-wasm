
class ExampleRunnable: Runnable {
    func run(input: String) -> String {
        return "hello " + input
    }
}

@_cdecl("init")
func `init`() {
    set(runnable: ExampleRunnable())
}