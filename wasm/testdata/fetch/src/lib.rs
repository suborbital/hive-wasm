use suborbital::runnable;
use suborbital::net;
use suborbital::util;

struct Fetch{}

impl runnable::Runnable for Fetch {
    fn run(&self, input: Vec<u8>) -> Option<Vec<u8>> {
        let url = util::to_string(input);
    
        let result = net::get(url.as_str());

        Some(result)
    }
}


// initialize the runner, do not edit below //
static RUNNABLE: &Fetch = &Fetch{};

#[no_mangle]
pub extern fn init() {
    runnable::set(RUNNABLE);
}
