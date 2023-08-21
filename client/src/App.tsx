import './fonts.css'

import Auth from './pages/Auth';
import Chat from './pages/Chat';
import { useState } from 'react';

function App() {
    const [token, setToken] = useState("");

    if (!token) {
        return <Auth setToken={setToken}></Auth>
    }

    return <Chat token={token}></Chat>

}

export default App;
