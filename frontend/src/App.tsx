import { useState } from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import { Greet, Launch } from "../wailsjs/go/main/App.js";
import { EventsOn, Quit } from "../wailsjs/runtime/runtime.js";
import { Button } from "@/components/ui/button"
import { ThemeProvider } from "@/components/theme-provider";
import { VersionCombobox } from '@/components/ui/version-combo';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { DownloadProgress } from './components/ui/download-progress';

function App() {
    const [resultText, setResultText] = useState("Launch");
    const [name, setName] = useState('');
    const [version, setVersion] = useState('');
    const updateResultText = (result: string) => setResultText(result);

    function launch() {
        updateResultText("Downloading...");
        Launch(version, name);
    }

    EventsOn("launching", (id) => {
        updateResultText("Launching Minecraft " + id)
        setTimeout(() => {
            Quit();
        }, 5000);
    });

    return (
        <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
            <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
                <Card className='w-full max-w-sm'>
                    <CardHeader>
                        <CardTitle className="text-2xl">Minecraft</CardTitle>
                        <CardDescription>
                            Select the Minecraft version you want to play and enter your username.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="grid w-full max-w-sm gap-2">
                            <div className="grid w-full gap-1">
                                <Label htmlFor='version-combobox'>Minecraft version</Label>
                                <VersionCombobox id='version-combobox' onSelected={setVersion} />
                            </div>
                            <div className="grid w-full gap-1">
                                <Label htmlFor="username">Your username</Label>
                                <Input placeholder="Type your username here." id="username" onChange={(ev) => setName(ev.target.value)} />
                            </div>
                            <div>
                                <Button onClick={launch} disabled={
                                    // disabled if version is empty or name is empty
                                    version === "" || name === ""
                                } className="w-full">{resultText}</Button>
                            </div>
                            <DownloadProgress />
                        </div>
                    </CardContent>
                </Card>
            </div>
        </ThemeProvider>
    )
}

export default App
