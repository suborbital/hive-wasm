
@_silgen_name("return_result_swift")
func return_result(result_pointer: UnsafeRawPointer, result_size: Int32, ident: Int32)
@_silgen_name("log_msg_swift")
func log_msg_swift(pointer: UnsafeRawPointer, size: Int32, level: Int32, ident: Int32)

// keep track of the current ident
var CURRENT_IDENT: Int32 = 0

// the Runnable instance currently being used
var RUNNABLE: Runnable = defaultRunnable()

// the protocol that users conform to to make their package a Runnable
public protocol Runnable {
    func run(input: String) -> String
}

// something to hold the Runnable's place until set is called
class defaultRunnable: Runnable {
    func run(input: String) -> String {
        return ""
    }
}

public func Set(runnable: Runnable) {
    RUNNABLE = runnable
}

public func LogInfo(msg: String) {
    let printCount = Int32(msg.utf8.count)

    let _ = msg.withCString( { (msgPtr) -> UInt in
        log_msg_swift(pointer: msgPtr, size: printCount, level: 3, ident: CURRENT_IDENT)
        return 0
    })
}

public func LogWarn(msg: String) {
    let printCount = Int32(msg.utf8.count)

    let _ = msg.withCString( { (msgPtr) -> UInt in
        log_msg_swift(pointer: msgPtr, size: printCount, level: 2, ident: CURRENT_IDENT)
        return 0
    })
}

public func LogErr(msg: String) {
    let printCount = Int32(msg.utf8.count)

    let _ = msg.withCString( { (msgPtr) -> UInt in
        log_msg_swift(pointer: msgPtr, size: printCount, level: 1, ident: CURRENT_IDENT)
        return 0
    })
}

@_cdecl("run_e")
func run_e(pointer: UnsafeRawPointer, size: Int32, ident: Int32) {
    CURRENT_IDENT = ident
    
    // convert the bytes to a string
    let typed: UnsafePointer<UInt8> = pointer.bindMemory(to: UInt8.self, capacity: Int(size))
    let inString = String(cString: typed)
    
    // call the user-provided run function
    let retString = RUNNABLE.run(input: inString)

    // convert the output to a usable pointer/size combo
    let count = Int32(retString.utf8.count)
    
    let _ = retString.withCString({ (retPtr) -> UInt in
        return_result(result_pointer: retPtr, result_size: count, ident: ident)
        return 0
    })
}

@_cdecl("allocate")
func allocate(size: Int) -> UnsafeMutableRawPointer {
  return UnsafeMutableRawPointer.allocate(byteCount: size, alignment: MemoryLayout<UInt8>.alignment)
}

@_cdecl("deallocate")
func deallocate(pointer: UnsafeRawPointer, size: Int) {
    let ptr: UnsafePointer<UInt8> = pointer.bindMemory(to: UInt8.self, capacity: Int(size))
    ptr.deallocate()
}