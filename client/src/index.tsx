import "./index.css";

import App from "./App";
import { CookiesProvider } from "react-cookie";
import React from "react";
import ReactDOM from "react-dom/client";

const root = ReactDOM.createRoot(
    document.getElementById("root") as HTMLElement
);
root.render(
    <React.StrictMode>
        <CookiesProvider
            defaultSetOptions={{
                path: "/",
                maxAge: 3600,
            }}
        >
            <App />
        </CookiesProvider>
    </React.StrictMode>
);
