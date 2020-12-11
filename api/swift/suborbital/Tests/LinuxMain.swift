import XCTest

import suborbitalTests

var tests = [XCTestCaseEntry]()
tests += suborbitalTests.allTests()
XCTMain(tests)
