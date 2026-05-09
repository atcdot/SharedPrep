import { BrowserRouter, Routes, Route } from "react-router-dom";
import { AuthProvider } from "@/lib/auth";
import { Layout } from "@/components/Layout";
import { RequireAuth } from "@/components/RequireAuth";
import { HomePage } from "@/pages/HomePage";
import { CreateEventPage } from "@/pages/CreateEventPage";
import { EventPage } from "@/pages/EventPage";
import { JoinPage } from "@/pages/JoinPage";
import { SettingsPage } from "@/pages/SettingsPage";

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Layout>
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/create" element={<RequireAuth><CreateEventPage /></RequireAuth>} />
            <Route path="/event/:shareCode" element={<EventPage />} />
            <Route path="/join/:shareCode" element={<JoinPage />} />
            <Route path="/settings" element={<RequireAuth><SettingsPage /></RequireAuth>} />
          </Routes>
        </Layout>
      </AuthProvider>
    </BrowserRouter>
  );
}
