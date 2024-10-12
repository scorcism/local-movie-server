import React, { useEffect, useState } from 'react';
import ReactPlayer from 'react-player';

const App: React.FC = () => {
  const [movies, setMovies] = useState<string[]>([]);
  const [selectedMovie, setSelectedMovie] = useState<string | null>(null);

  useEffect(() => {
    const fetchMovies = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/movies');
        const data = await response.text();
        const movieList = data.split('\n').filter((movie) => movie.trim() !== '');
        setMovies(movieList);
      } catch (error) {
        console.error('Error fetching movies:', error);
      }
    };

    fetchMovies();
  }, []);

  const selectMovie = (movie: string) => {
    setSelectedMovie(movie);
  };

  return (
    <div className="App">
      <h1>My Movie Collection</h1>
      <div>
        <h2>Available Movies</h2>
        <ul>
          {movies.map((movie, index) => (
            <li key={index}>
              <button onClick={() => selectMovie(movie)}>{movie}</button>
            </li>
          ))}
        </ul>
      </div>
      {selectedMovie && (
        <div>
          <h2>Now Playing: {selectedMovie}</h2>
          <ReactPlayer
            url={`http://localhost:8080/api/stream?name=${encodeURIComponent(selectedMovie)}`}
            controls={true}
            width="100%"
            height="100%"
          />
        </div>
      )}
    </div>
  );
};

export default App;
