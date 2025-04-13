"use client"

import * as React from "react"
import { Check, ChevronsUpDown } from "lucide-react"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
    Command,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
} from "@/components/ui/command"
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover"
import { GetVersions } from "../../../wailsjs/go/main/App.js";

interface VersionComboboxProps {
    id: string;
    className?: string;
    onSelected?: (value: string) => void;
}

export function VersionCombobox({ id, className, onSelected }: VersionComboboxProps) {
    const [open, setOpen] = React.useState(false)
    const [value, setValue] = React.useState("")
    const [versions, setVersions] = React.useState([""])

    React.useEffect(() => {
        GetVersions().then((versions) => setVersions(versions))
    }, [])


    return (
        <div id={id} className={cn("w-full", className)}>
            <Popover open={open} onOpenChange={setOpen}>
                <PopoverTrigger asChild>
                    <Button
                        variant="outline"
                        role="combobox"
                        aria-expanded={open}
                        className="w-[200px] justify-between"
                    >
                        {value
                            ? versions.find((version) => version === value)
                            : "Select version..."}
                        <ChevronsUpDown className="opacity-50" />
                    </Button>
                </PopoverTrigger>
                <PopoverContent className="w-[200px] p-0">
                    <Command>
                        <CommandInput placeholder="Search version..." className="h-9" />
                        <CommandList>
                            <CommandEmpty>No versions found</CommandEmpty>
                            <CommandGroup>
                                {versions.map((version) => (
                                    <CommandItem
                                        key={version}
                                        value={version}
                                        onSelect={(currentValue) => {
                                            setValue(currentValue === value ? "" : currentValue)
                                            onSelected?.(currentValue === value ? "" : currentValue)
                                            setOpen(false)
                                        }}
                                    >
                                        {version}
                                        <Check
                                            className={cn(
                                                "ml-auto",
                                                value === version ? "opacity-100" : "opacity-0"
                                            )}
                                        />
                                    </CommandItem>
                                ))}
                            </CommandGroup>
                        </CommandList>
                    </Command>
                </PopoverContent>
            </Popover>
        </div>
    )
}
