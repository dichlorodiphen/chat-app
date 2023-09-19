import "./fonts.css";

import Auth from "./pages/Auth";
import Chat from "./pages/Chat";
import { useCookies } from "react-cookie";

function App() {
    const [cookies, setCookie, removeCookie] = useCookies(["token"]);

    if (!cookies["token"]) {
        return <Auth></Auth>;
    }

    return <Chat token={cookies["token"]}></Chat>;
}

export default App;
