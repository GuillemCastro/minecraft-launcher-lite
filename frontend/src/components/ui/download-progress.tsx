import * as React from "react"

import { Progress } from "@/components/ui/progress"
import { EventsOn } from "../../../wailsjs/runtime/runtime"

export function DownloadProgress() {
    const [progress, setProgress] = React.useState(0)
    const [hidden, setHidden] = React.useState(true)
    const [max, setMax] = React.useState<number>()

    EventsOn("downloadStart", (total) => {
        setProgress(0)
        setHidden(false)
        setMax(total)
    })
    EventsOn("downloadProgress", (progress, left) => {
        if (progress !== undefined && max !== undefined) {
            setProgress((progress / max) * 100)
        }
    })
    EventsOn("downloadFinished", () => {
        setProgress(100)
        setHidden(true)
    })

    return <Progress hidden={hidden} value={progress} max={100} className="w-full" />
}
