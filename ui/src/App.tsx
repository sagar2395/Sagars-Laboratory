
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Layout } from './components/Layout';
import { DashboardView } from './views/DashboardView';
import { ScenariosView } from './views/ScenariosView';
import { PlatformView } from './views/PlatformView';
import { AppsView } from './views/AppsView';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<DashboardView />} />
          <Route path="scenarios" element={<ScenariosView />} />
          <Route path="platform" element={<PlatformView />} />
          <Route path="apps" element={<AppsView />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
