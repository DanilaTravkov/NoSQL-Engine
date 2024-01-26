import './globals.css';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Login, CreateUser } from './pages';
import AuthLayout from './components/auth/AuthLayout';

function App() {
  return (
    <Router>
      <main className='h-screen flex bg-yellow-200'>
        <Routes>
          <Route element={<AuthLayout/>}>
            <Route path="login" element={<Login />} />
            <Route path="create" element={<CreateUser />} />
          </Route>
          <Route path="*" element={<NotFound />} />
        </Routes>
      </main>
    </Router>
  );
}

function NotFound() {
  return <div className="flex flex-1 justify-center items-center flex-col py-10">
    404 Not Found
    </div>;
}

export default App;
