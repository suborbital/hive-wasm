use suborbital::runnable;
use suborbital::log;

struct Log{}

impl runnable::Runnable for Log {
    fn run(&self, input: Vec<u8>) -> Option<Vec<u8>> {
        let in_string = String::from_utf8(input).unwrap();

        log::info(in_string.as_str());
    
        Some(String::from("hello").as_bytes().to_vec())
    }
}


// initialize the runner, do not edit below //
static RUNNABLE: &Log = &Log{};

#[no_mangle]
pub extern fn init() {
    runnable::set(RUNNABLE);
}
