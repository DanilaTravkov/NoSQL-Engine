import './globals.css';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Login, CreateUser } from './pages';
import AuthLayout from './components/auth/AuthLayout';
import Home from './pages/Home';

function App() {
  return (
    <Router>
      <main className='h-screen flex bg-green-600'>
        <Routes>
          {/* auth routes */}
          <Route element={<AuthLayout/>}>
            <Route path="/login" element={<Login />} />
            <Route path="/create" element={<CreateUser />} />
          </Route>
          {/* other routes */}
          <Route path='/' element={<Home />}/>
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
