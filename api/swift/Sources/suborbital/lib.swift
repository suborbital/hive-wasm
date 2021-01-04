
@_silgen_name("return_result_swift")
func return_result(result_pointer: UnsafeRawPointer, result_size: Int32, ident: Int32)
@_silgen_name("log_msg_swift")
func log_msg_swift(pointer: UnsafeRawPointer, size: Int32, level: Int32, ident: Int32)
@_silgen_name("cache_set_swift")
func cache_set_swift(key_pointer: UnsafeRawPointer, key_size: Int32, value_pointer: UnsafeRawPointer, value_size: Int32, ttl: Int32, ident: Int32) -> Int32
@_silgen_name("cache_get_swift")
func cache_get_swift(key_pointer: UnsafeRawPointer, key_size: Int32, dest_pointer: UnsafeRawPointer, dest_max_size: Int32, ident: Int32) -> Int32

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

public func CacheSet(key: String, value: String, ttl: Int) {
    let keyCount = Int32(key.utf8.count)
    let valCount = Int32(value.utf8.count)
    
    let _ = key.withCString({ (keyPtr) -> UInt in
        let _ = value.withCString({ (valPtr) -> UInt in
            let _ = cache_set_swift(key_pointer: keyPtr, key_size: keyCount, value_pointer: valPtr, value_size: valCount, ttl: Int32(ttl), ident: CURRENT_IDENT)
            return 0
        })

        return 0
    })
}

public func CacheGet(key: String) -> String {
    let keyCount = Int32(key.utf8.count)
    
    var maxSize: Int32 = 256000
    var retVal = ""

    let _ = key.withCString({ (keyPtr) -> UInt in
        while true {
            let ptr = allocate(size: Int(maxSize)) 
            let resultSize = cache_get_swift(key_pointer: keyPtr, key_size: keyCount, dest_pointer: ptr, dest_max_size: maxSize, ident: CURRENT_IDENT)

            if resultSize < 0 {
                retVal = "failed to get from cache"
                break
            } else if resultSize > maxSize {
                maxSize *= 2
            } else {
                let typed: UnsafeMutablePointer<UInt8> = ptr.bindMemory(to: UInt8.self, capacity: Int(resultSize))
                retVal = String(cString: typed)
                break
            }
        }

        return 0
    })
    
    return retVal
}

public func LogInfo(msg: String) {
    log(msg: msg, level: 3)
}

public func LogWarn(msg: String) {
    log(msg: msg, level: 2)
}

public func LogErr(msg: String) {
    log(msg: msg, level: 1)
}

func log(msg: String, level: Int32) {
    let printCount = Int32(msg.utf8.count)

    let _ = msg.withCString( { (msgPtr) -> UInt in
        log_msg_swift(pointer: msgPtr, size: printCount, level: level, ident: CURRENT_IDENT)
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