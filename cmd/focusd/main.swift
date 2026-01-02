import Foundation
import AppKit
import CoreGraphics

struct FocusOutput: Codable {
    let ts_ms: Int64
    let app_name: String
    let bundle_id: String
    let pid: Int
    let window_title: String
}

guard let app = NSWorkspace.shared.frontmostApplication else {
    exit(1)
}

let tsMs = Int64(Date().timeIntervalSince1970 * 1000)
let appName = app.localizedName ?? ""
let bundleId = app.bundleIdentifier ?? ""
let pid = Int(app.processIdentifier)

var windowTitle = ""

if let windowList = CGWindowListCopyWindowInfo([.optionOnScreenOnly, .excludeDesktopElements], kCGNullWindowID) as? [[String: Any]] {
    for window in windowList {
        if let windowOwnerPID = window[kCGWindowOwnerPID as String] as? Int, windowOwnerPID == pid {
            // Get the window name (title)
            if let name = window[kCGWindowName as String] as? String, !name.isEmpty {
                windowTitle = name
                break // Assume the first one found is the main one
            }
        }
    }
}

let output = FocusOutput(ts_ms: tsMs, app_name: appName, bundle_id: bundleId, pid: pid, window_title: windowTitle)
let encoder = JSONEncoder()
if let data = try? encoder.encode(output) {
    if let json = String(data: data, encoding: .utf8) {
        print(json)
    }
}
